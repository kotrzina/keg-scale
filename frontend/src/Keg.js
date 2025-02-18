import {Alert, Col, Offcanvas, Row} from "react-bootstrap";
import Button from "react-bootstrap/Button";
import React from "react";
import {buildUrl} from "./Api";
import useApiPassword from "./useApiPassword";
import PasswordBox from "./PasswordBox";

function Keg(props) {

    const [showError, setShowError] = React.useState(false)
    const [apiPassword, isApiReady] = useApiPassword()

    const kegs = [0, 10, 15, 20, 30, 50]

    async function switchKeg(size) {

        const request = new Request(buildUrl("/api/pub/active_keg"), {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": apiPassword,
            },
            body: JSON.stringify({keg: size}),
        });

        const response = await fetch(request)
        if (response.status === 200) {
            props.refresh()
            props.setShowCanvas(false)
            setShowError(false)
        } else {
            setShowError(true)
        }
    }

    return (
        <Offcanvas show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Naražená bečka</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>
                <Row hidden={!isApiReady}>
                    <Alert hidden={!showError} variant={"danger"}>
                        Chyba! Zkus to prosim pozdeji.
                    </Alert>

                    <Col md={12}>
                        {kegs.map((keg) => {
                                return (
                                    <Button
                                        key={keg}
                                        className={"m-1"}
                                        variant={props.keg === keg ? "success" : "primary"}
                                        size={"lg"}
                                        onClick={() => {
                                            void switchKeg(keg)
                                        }}>{keg}</Button>
                                )
                            }
                        )}
                    </Col>
                </Row>

                <PasswordBox/>

            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Keg;