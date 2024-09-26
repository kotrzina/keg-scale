import {Alert, Col, Offcanvas, Row} from "react-bootstrap";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import React, {useEffect} from "react";
import {buildUrl} from "./Api";

function Keg(props) {

    const [showError, setShowError] = React.useState(false)
    const [password, setPassword] = React.useState("")

    const kegs = [10, 15, 20, 30, 50]

    async function switchKeg(size) {

        const request = new Request(buildUrl("/api/pub/active_keg"), {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": password,
            },
            body: JSON.stringify({keg: size}),
        });

        const response = await fetch(request)
        if (response.status === 200) {
            props.refresh()
            props.setShowCanvas(false)
            setShowError(false)
            localStorage.setItem("password", password)
        } else {
            setShowError(true)
        }
    }

    useEffect(() => {
        if (password !== "") {
            return
        }

        const storedPassword = localStorage.getItem("password")
        if (storedPassword !== null && storedPassword !== "") {
            setPassword(storedPassword)
        }
    }, [password]);

    return (
        <Offcanvas show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Naražená bečka</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>
                <Row>

                    <Alert hidden={!showError} variant={"danger"}>
                        Chyba! Asi špatné heslo.
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

                    <div className={"mt-3"}></div>
                    <Form className="d-flex">
                        <Form.Control
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            type="password"
                            placeholder="Heslo"
                            className="me-2"
                            aria-label="Heslo"
                        />
                        <Form.Text className="text-muted">
                            <code>heslo</code>
                        </Form.Text>
                    </Form>
                </Row>


            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Keg;