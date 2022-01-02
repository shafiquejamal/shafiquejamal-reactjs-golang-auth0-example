interface Config {
    auth0_client_id: string,
    auth0_domain: string,
    auth0_redirect_uri: string,
}

const config: Config = {
    auth0_client_id: (process.env.REACT_APP_AUTH0_CLIENT_ID as string),
    auth0_domain: (process.env.REACT_APP_AUTH0_DOMAIN as string),
    auth0_redirect_uri: (process.env.REACT_APP_AUTH0_REDIRECT_URI as string),
}

export default config;