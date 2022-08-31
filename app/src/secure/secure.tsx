import { useAuth0 } from '@auth0/auth0-react';
import { useEffect, useState } from "react";
import React from 'react';
import { LogoutButton } from '../components/LogoutButton';
import { Ping } from '../ping/Ping';

export const Secure = () => {
  const { isLoading, user, getAccessTokenSilently } = useAuth0();
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

  if (isLoading) {
    return <div>Loading</div>;
  }

  return (
    <div>
      <h1>Secure area!</h1>
      <p>{JSON.stringify(user)}</p>
      <p>accessToken:{accessToken}</p>
      <Ping />
      <LogoutButton />
    </div>
  );
}