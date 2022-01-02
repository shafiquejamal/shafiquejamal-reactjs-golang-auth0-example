import { useAuth0 } from '@auth0/auth0-react';
import React from 'react';
import { LoginButton } from '../components/LoginButton';

export const Unsecured = () => {
  const { user } = useAuth0();

  if (user && user.email_verified === false) {
      return <div>You need to verify your email!</div>;
  }

  return <div>You are not authenticated! <LoginButton /></div>;
}