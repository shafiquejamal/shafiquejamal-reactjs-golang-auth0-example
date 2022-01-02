import { useRoutes } from 'react-router';
import { useAuth0 } from '@auth0/auth0-react';
import { Secure } from './secure/secure';
import { Unsecured } from './unsecured/unsecured';
import Index from './index/index';
import { useEffect } from 'react';

const Routes = () => {
    const { isAuthenticated, user, getAccessTokenSilently } = useAuth0();

    const canAccess = isAuthenticated && user.email_verified === true;

    useEffect(() => {
      async function fetchData() {
        if (isAuthenticated && user.email_verified === false) { 
          await getAccessTokenSilently({ignoreCache: true})
        }
      }

      fetchData();
    }, [user, isAuthenticated, getAccessTokenSilently]);

    const routes = [
        { path: '/secure', element: (canAccess ? <Secure /> : <Unsecured />) },
        { path: '/', element: <Index /> },
      ];
      
    const routing = useRoutes(routes);

    return routing;
}

export default Routes;
  