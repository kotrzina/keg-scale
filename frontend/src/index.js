import 'bootstrap/dist/css/bootstrap.min.css';
import * as React from "react";
import {createRoot} from "react-dom/client";
import {
    createBrowserRouter,
    RouterProvider,
} from "react-router-dom";
import Dashboard from "./Dashboard";

const router = createBrowserRouter([
    {
        path: "/",
        element: (
            <div><Dashboard/></div>
        ),
    },
    {
        path: "about",
        element: <div>About</div>,
    },
]);

createRoot(document.getElementById("root")).render(
    <RouterProvider router={router}/>
);