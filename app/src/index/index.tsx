import { useAuth0 } from "@auth0/auth0-react";
import { useEffect, useState } from "react";
import { LoginButton } from "../components/LoginButton";
import { LogoutButton } from "../components/LogoutButton";

function Index() {
  const { isAuthenticated, user, getAccessTokenSilently } = useAuth0();
  const [accessToken, setAccessToken] = useState('');
 
  useEffect(() => {
    (async () => {
      try {
        const token = await getAccessTokenSilently();
        setAccessToken(token)
      } catch (e) {
        console.error(e);
      }
    })();
  }, [getAccessTokenSilently, accessToken]);

  return (<div>
    {
      isAuthenticated ?
        <div>
          <p>{user?.email}</p>
          <p>Access token is: {accessToken}</p>
          <LogoutButton />
          
        </div> :
        <LoginButton />
    }
  </div>
  );
}

export default Index;
