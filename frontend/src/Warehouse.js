import {Alert, Col, Offcanvas, Row, Table} from "react-bootstrap";
import Form from "react-bootstrap/Form";
import React, {useEffect} from "react";
import WarehouseKeg from "./WarehouseKeg";


function Warehouse(props) {

    const [showError, setShowError] = React.useState(false)
    const [password, setPassword] = React.useState("")

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
                <Offcanvas.Title>Sklad</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>

                <Row>
                    <Alert hidden={!showError} variant={"danger"}>
                        Chyba! Asi špatné heslo.
                    </Alert>

                    <Col md={12}>
                        <Table bordered={false} align={"center"}>
                            <tbody>
                            {props.warehouse.map((keg) => {
                                return (
                                    <WarehouseKeg
                                        key={keg.keg}
                                        keg={keg}
                                        refresh={props.refresh}
                                        password={password}
                                        setShowError={setShowError}
                                    />
                                )
                            })}

                            </tbody>
                        </Table>
                    </Col>
                </Row>

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
            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Warehouse;