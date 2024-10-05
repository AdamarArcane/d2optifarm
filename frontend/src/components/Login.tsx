import React from 'react';

// Import Button component from Shadcn/UI
import { Button } from './ui/button'

const Login: React.FC = () => {
    const handleLogin = () => {
        window.location.href = 'https://gannet-sweeping-frequently.ngrok-free.app/v1/auth/login';
    };

    return (
        <div className="login-container" >
            <h2>Welcome to the Destiny 2 Companion App </h2>
            < Button onClick={handleLogin} > Login with Bungie </Button>
        </div>
    );
};

export default Login;