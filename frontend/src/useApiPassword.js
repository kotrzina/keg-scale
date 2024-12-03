import {useState, useEffect, useCallback} from 'react';
import {buildUrl} from "./Api";

export default function useApiPassword() {

    const STORAGE_KEY = "password";
    const [pwd, setPwd] = useState("");
    const [isOk, setIsOk] = useState(false);

    useEffect(() => {
        if (pwd !== "") {
            return
        }

        if (isOk) {
            return
        }

        const storedPassword = localStorage.getItem("password")
        if (storedPassword !== null && storedPassword !== "") {
            fetch(buildUrl("/api/check/password"), {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": storedPassword,
                },
            }).then(response => {
                if (response.status === 204) {
                    setPwd(storedPassword)
                    setIsOk(true)
                }
            })
        }
    }, [pwd, isOk]);

    const checkPassword = useCallback(async (password) => {
        if (password === "") {
            return
        }

        if (isOk) {
            return
        }

        const response = await fetch(buildUrl("/api/check/password"), {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Authorization": password,
            },
        })

        if (response.status === 204) {
            localStorage.setItem(STORAGE_KEY, password)
            window.location.reload() // reload the page to propagate the password
        } else {
            localStorage.removeItem(STORAGE_KEY)
        }
    }, [isOk]);

    function changePassword(password) {
        setPwd(password)
        void checkPassword(password)
    }

    return [pwd, isOk, changePassword];
}