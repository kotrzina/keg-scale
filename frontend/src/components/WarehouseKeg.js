import Button from "react-bootstrap/Button";
import React from "react";
import { buildUrl } from "../lib/Api";
import { useAuth } from "../contexts/AuthContext";

function WarehouseKeg(props) {

    const { password } = useAuth();

    async function onKegChange(way) {
        const request = new Request(buildUrl("/api/scale/warehouse"), {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": password,
            },
            body: JSON.stringify({
                keg: props.keg.keg,
                way: way,
            }),
        });

        const response = await fetch(request)
        if (response.status === 200) {
            props.refresh()
            props.setShowError(false)
        } else {
            props.setShowError(true)
        }
    }

    return (
        <tr>
            <td style={{ textAlign: "center" }}>
                <Button variant={"info"}
                    size={"lg"}
                    onClick={() => onKegChange("down")}
                >➖</Button>
            </td>
            <td style={{ textAlign: "center" }}>
                <Button variant={"outline-danger"} size={"lg"}>{props.keg.keg}&nbsp;l</Button>
                &nbsp;&nbsp;&nbsp;✖️&nbsp;&nbsp;&nbsp;
                <Button variant={"outline-secondary"} size={"lg"}>{props.keg.amount}</Button></td>
            <td style={{ textAlign: "center" }}>
                <Button variant={"info"}
                    size={"lg"}
                    onClick={() => onKegChange("up")}
                >➕</Button>
            </td>
        </tr>
    );
}

export default WarehouseKeg;
