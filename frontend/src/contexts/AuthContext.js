import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { buildUrl } from '../lib/Api';

const AuthContext = createContext(null);

const STORAGE_KEY = "password";

export function AuthProvider({ children }) {
    const [password, setPassword] = useState("");
    const [isAuthenticated, setIsAuthenticated] = useState(false);

    useEffect(() => {
        if (password !== "" || isAuthenticated) {
            return;
        }

        const storedPassword = localStorage.getItem(STORAGE_KEY);
        if (storedPassword !== null && storedPassword !== "") {
            fetch(buildUrl("/api/check/password"), {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": storedPassword,
                },
            }).then(response => {
                if (response.status === 204) {
                    setPassword(storedPassword);
                    setIsAuthenticated(true);
                }
            }).catch(() => {
                setIsAuthenticated(false);
            });
        }
    }, [password, isAuthenticated]);

    const login = useCallback(async (newPassword) => {
        if (newPassword === "" || isAuthenticated) {
            return;
        }

        try {
            const response = await fetch(buildUrl("/api/check/password"), {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": newPassword,
                },
            });

            if (response.status === 204) {
                localStorage.setItem(STORAGE_KEY, newPassword);
                setPassword(newPassword);
                setIsAuthenticated(true);
            } else {
                localStorage.removeItem(STORAGE_KEY);
            }
        } catch {
            setIsAuthenticated(false);
        }
    }, [isAuthenticated]);

    const logout = useCallback(() => {
        localStorage.removeItem(STORAGE_KEY);
        setPassword("");
        setIsAuthenticated(false);
    }, []);

    return (
        <AuthContext.Provider value={{ password, isAuthenticated, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
