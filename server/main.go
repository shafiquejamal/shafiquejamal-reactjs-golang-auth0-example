package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/shafiquejamal/reactjs-golang-starter/auth"
	"github.com/shafiquejamal/reactjs-golang-starter/cf"
	"github.com/spf13/viper"

	"github.com/gorilla/mux"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {
	// make config values available
	viper.SetConfigName("config")
	viper.AddConfigPath("./cf")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	var configuration cf.Configuration
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(configuration.AwsRegion),
		config.WithSharedCredentialsFiles([]string{".aws/credentials"}),
		config.WithSharedConfigProfile(configuration.AwsProfile))
	if err != nil {
		panic(err)
	}
	svc := dynamodb.NewFromConfig(cfg)

	r := mux.NewRouter()
	apiPrefix := configuration.ApiPrefix
	r.Handle(apiPrefix+"/ping", auth.RequireAuthentication(auth.AllowAllAuthorizationStrategy, auth.OAuthUserIdentityFetcher(configuration.AuthServerUserInfoEndpoint), userDataFetcher(configuration.DdbUserAccessPolicyTableName, configuration.DdbAccessPolicyTableName, configuration.DdbPolicyGroupTableName, svc))(pingHandler())).Methods("GET")
	r.Handle(apiPrefix+"/pong", auth.RequireAuthentication(auth.PolicyAuthorizationStrategy(apiPrefix), auth.OAuthUserIdentityFetcher(configuration.AuthServerUserInfoEndpoint), userDataFetcher(configuration.DdbUserAccessPolicyTableName, configuration.DdbAccessPolicyTableName, configuration.DdbPolicyGroupTableName, svc))(pingHandler())).Methods("GET")
	r.Handle(apiPrefix+"/pung", auth.RequireAuthentication(auth.PolicyAuthorizationStrategy(apiPrefix), auth.OAuthUserIdentityFetcher(configuration.AuthServerUserInfoEndpoint), userDataFetcher(configuration.DdbUserAccessPolicyTableName, configuration.DdbAccessPolicyTableName, configuration.DdbPolicyGroupTableName, svc))(pingHandler())).Methods("GET")
	r.Handle(apiPrefix+"/pang", auth.RequireAuthentication(auth.PolicyAuthorizationStrategy(apiPrefix), auth.OAuthUserIdentityFetcher(configuration.AuthServerUserInfoEndpoint), userDataFetcher(configuration.DdbUserAccessPolicyTableName, configuration.DdbAccessPolicyTableName, configuration.DdbPolicyGroupTableName, svc))(pingHandler())).Methods("GET")

	spa := spaHandler{staticPath: "../app/build", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa).Methods("GET")

	port := "8090"
	log.Println("Server starting at port " + port + "...")

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func pingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := (r.Context().Value("User")).(auth.User)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(user.Identity.Email))
	}
}

type UserPolicyNames struct {
	PolicyNames  []string `dynamodbav:"access_policies"`
	UserId       string   `dynamodbav:"user_id"`
	PolicyGroups []string `dynamodbav:"policy_groups"`
}

type PolicyGroup struct {
	Name        string
	PolicyNames []string `dynamodbav:"policy_names"`
}

type PermsWithMeta struct {
	Name        string
	Created_at  int
	Updated_at  int
	Description string
	Permissions auth.Permission
}

func userDataFetcher(userAccessPoliciesTableName, accessPoliciesTableName, policyGroupsTableName string, svc *dynamodb.Client) func(userI *auth.UserIdentity, u *auth.User, w *http.ResponseWriter) error {
	return func(userI *auth.UserIdentity, u *auth.User, w *http.ResponseWriter) error {

		if userI.UserId == "" {
			return errors.New("UserId is empty")
		}

		// Get names of policies attached directly
		directlyAttachedPoliciesResult, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
			TableName: aws.String(userAccessPoliciesTableName),
			Key: map[string]types.AttributeValue{
				"user_id": &types.AttributeValueMemberS{Value: userI.UserId},
			},
		})
		if err != nil {
			return errors.New(fmt.Sprintf("Got error calling GetItem: %s", err))
		}
		userPolicyNames := UserPolicyNames{}
		err = attributevalue.UnmarshalMap(directlyAttachedPoliciesResult.Item, &userPolicyNames)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}
		policyNames := userPolicyNames.PolicyNames

		// Get policy names from groups
		policyGroupsForQuery := []map[string]types.AttributeValue{}
		for _, group := range userPolicyNames.PolicyGroups {
			policyGroupsForQuery = append(policyGroupsForQuery, map[string]types.AttributeValue{
				"name": &types.AttributeValueMemberS{Value: group},
			})
		}
		policiesFromGroupsResult, err := svc.BatchGetItem(context.TODO(), &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				policyGroupsTableName: {
					Keys: policyGroupsForQuery,
				},
			},
		})
		if err != nil {
			return errors.New(fmt.Sprintf("Got error calling BatchGetItem: %s", err))
		}
		for _, table := range policiesFromGroupsResult.Responses {
			for _, item := range table {
				var policyGroup PolicyGroup
				err = attributevalue.UnmarshalMap(item, &policyGroup)

				if err != nil {
					return errors.New(fmt.Sprintf("failed to unmarshall place from dynamodb response, err: %s", err))
				}
				policyNames = append(policyNames, policyGroup.PolicyNames...)
			}
		}
		toSet(&policyNames)

		avs := []map[string]types.AttributeValue{}
		for _, accessPolicy := range policyNames {
			avs = append(avs, map[string]types.AttributeValue{
				"name": &types.AttributeValueMemberS{Value: accessPolicy},
			})
		}
		permissionsResult, err := svc.BatchGetItem(context.TODO(), &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				accessPoliciesTableName: {
					Keys: avs,
				},
			},
		})
		if err != nil {
			return errors.New(fmt.Sprintf("err2: %v", permissionsResult))
		}

		authPermissions := []auth.Permission{}
		for _, table := range permissionsResult.Responses {
			for _, item := range table {
				permission := PermsWithMeta{}
				err = attributevalue.UnmarshalMap(item, &permission)

				if err != nil {
					return errors.New(fmt.Sprintf("failed to unmarshall place from dynamodb response, err: %s", err))
				}
				authPermissions = append(authPermissions, permission.Permissions)
			}
		}

		(*u).Identity = *userI
		(*u).Permissions = authPermissions
		return nil
	}
}

func toSet(slice *[]string) {
	processed := map[string]struct{}{}
	w := 0
	for _, s := range *slice {
		if _, exists := processed[s]; !exists {
			// If this city has not been seen yet, add it to the list
			processed[s] = struct{}{}
			(*slice)[w] = s
			w++
		}
	}
	*slice = (*slice)[:w]
}
