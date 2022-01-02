import { useAuth0 } from '@auth0/auth0-react';
import React from 'react';
import { LogoutButton } from '../components/LogoutButton';
import { Ping } from '../ping/Ping';

export const Secure = () => {
  const { isLoading, user } = useAuth0();

  if (isLoading) {
    return <div>Loading</div>;
  }

  return (
    <div>
      <h1>Secure area!</h1>
      <p>{JSON.stringify(user)}</p>
      <Ping />
      <LogoutButton />
    </div>
  );
}