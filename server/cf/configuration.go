package cf

type Configuration struct {
	ClientSecret                 string
	AuthServerUserInfoEndpoint   string
	ApiPrefix                    string
	DdbAccessKeyId               string
	DdbSecretAccessKey           string
	DdbUserAccessPolicyTableName string
	DdbAccessPolicyTableName     string
	DdbPolicyGroupTableName      string
	AwsRegion                    string
	AwsProfile                   string
}
