import React from 'react';
import ReactDOM from 'react-dom';
import reportWebVitals from './reportWebVitals';
import { BrowserRouter as Router } from 'react-router-dom';
import { Auth0Provider } from '@auth0/auth0-react';
import Routes from './routes';
import config from './config'

ReactDOM.render(
  <React.StrictMode>
    <Router>
      <Auth0Provider
        domain={config.auth0_domain as string}
        clientId={config.auth0_client_id as string}
        redirectUri={config.auth0_redirect_uri as string}
        useRefreshTokens={true}
        cacheLocation="localstorage"
      >
        <Routes />
      </Auth0Provider>
    </Router>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
