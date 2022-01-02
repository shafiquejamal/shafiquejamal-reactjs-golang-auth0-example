import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { useAuth0 } from '@auth0/auth0-react';

export const Ping = () => {
    const { getAccessTokenSilently } = useAuth0();

    const [notification, setNotification] = useState('');
    const [accessToken, setAccessToken] = useState('');

    useEffect(() => {
        (async () => {
            try {
                const token = await getAccessTokenSilently();
                axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
                setAccessToken(token);
            } catch (e) {
                console.error(e);
            }
        })();
    }, [getAccessTokenSilently, accessToken]);

    const handlePing = async () => {
        try {
            const response = await axios.get('/api/ping');
            setNotification(`Successful ping with response: ${response.data}`);
        } catch (e) {
            setNotification('Failed to ping');
        }

        setTimeout(() => setNotification(''), 2000);
    }


    return (
        <div>
            <div>
                <p>{notification}</p>

                <button onClick={handlePing}>Ping</button>
            </div>
        </div>
    );
}