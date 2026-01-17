import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { buildUrl } from '../lib/Api';
import { useAuth } from './AuthContext';

const DashboardContext = createContext(null);

const defaultScale = {
    scale: {
        is_ok: false,
        beers_left: 0,
        beers_total: 0,
        last_weight: 0.0,
        last_weight_formated: "0.0",
        last_at: "0",
        last_at_duration: "0",
        rssi: 0,
        last_update: 0,
        last_update_duration: 0,
        pub: {
            is_open: false,
            opened_at: 0,
            closed_at: 0,
        },
        active_keg: 0,
        is_low: false,
        warehouse: [
            { "keg": 10, "amount": 0 },
            { "keg": 15, "amount": 0 },
            { "keg": 20, "amount": 0 },
            { "keg": 30, "amount": 0 },
            { "keg": 50, "amount": 0 }
        ],
        warehouse_beer_left: 0,
        bank_balance: {
            balance: "0"
        },
        bank_transactions: [
            {
                date: "",
                amount: "",
                account_name: "",
                bank_name: "",
                bank_code: "",
                recipient_message: "",
                comment: "",
                user_identification: "",
            }
        ],
    },
};

export function DashboardProvider({ children }) {
    const { password } = useAuth();
    const [data, setData] = useState(defaultScale);
    const [showKeg, setShowKeg] = useState(false);
    const [showBank, setShowBank] = useState(false);
    const [showWarehouse, setShowWarehouse] = useState(false);
    const [showChat, setShowChat] = useState(false);
    const [isLoading, setIsLoading] = useState(false);

    const refresh = useCallback(async () => {
        setIsLoading(true);
        try {
            const request = new Request(buildUrl("/api/scale/dashboard"), {
                method: "GET",
                headers: {
                    "Authorization": password,
                },
            });
            const res = await fetch(request);
            const responseData = await res.json();
            setData(responseData);
        } catch {
            setData(defaultScale);
        }
        setIsLoading(false);
    }, [password]);

    useEffect(() => {
        void refresh();

        window.addEventListener("focus", refresh);
        const interval = setInterval(() => {
            void refresh();
        }, 10000);

        return () => {
            window.removeEventListener("focus", refresh);
            clearInterval(interval);
        };
    }, [refresh]);

    return (
        <DashboardContext.Provider value={{
            data,
            isLoading,
            refresh,
            showKeg,
            setShowKeg,
            showBank,
            setShowBank,
            showWarehouse,
            setShowWarehouse,
            showChat,
            setShowChat,
        }}>
            {children}
        </DashboardContext.Provider>
    );
}

export function useDashboard() {
    const context = useContext(DashboardContext);
    if (!context) {
        throw new Error('useDashboard must be used within a DashboardProvider');
    }
    return context;
}
