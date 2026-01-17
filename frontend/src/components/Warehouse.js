import { Alert, Col, Offcanvas, Row, Table } from "react-bootstrap";
import React from "react";
import WarehouseKeg from "./WarehouseKeg";
import { useAuth } from "../contexts/AuthContext";
import PasswordBox from "./PasswordBox";


function Warehouse(props) {

    const { isAuthenticated } = useAuth();
    const [showError, setShowError] = React.useState(false)

    return (
        <Offcanvas show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Sklad</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>

                <Row hidden={!isAuthenticated}>
                    <Alert hidden={!showError} variant={"danger"}>
                        Chyba! Zkus to prosím později.
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
                                        setShowError={setShowError}
                                    />
                                )
                            })}
                            </tbody>
                        </Table>
                    </Col>
                </Row>

                <PasswordBox />
            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Warehouse;
