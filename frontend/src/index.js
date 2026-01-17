import 'bootstrap/dist/css/bootstrap.min.css';
import './index.css';
import * as React from "react";
import { createRoot } from "react-dom/client";
import {
    createBrowserRouter,
    RouterProvider,
} from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Layout />,
        children: [
            {
                index: true,
                element: <Dashboard />,
            },
        ],
    },
]);

createRoot(document.getElementById("root")).render(
    <AuthProvider>
        <RouterProvider router={router} />
    </AuthProvider>
);
