import {Col, Offcanvas, Row} from "react-bootstrap";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import React from "react";


function Warehouse(props) {
    return (
        <Offcanvas show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Sklad</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>
                <Form className="d-flex">
                    <Form.Control
                        type="number"
                        placeholder="Kód"
                        className="me-2"
                        aria-label="Search"
                    />
                    <Button variant="outline-success">Ok</Button>
                </Form>
                <Row>
                    <Col md={12}>
                        <p>
                            Zde se bude dat upratovat sklad.
                        </p>
                    </Col>
                </Row>

            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Warehouse;