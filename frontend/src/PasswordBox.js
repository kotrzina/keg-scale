import useApiPassword from "./useApiPassword";
import Form from "react-bootstrap/Form";
import {Col, Row} from "react-bootstrap";
import React from "react";

function PasswordBox(props) {

    const [apiPassword, isApiPasswordOk, changeApiPassword] = useApiPassword()

    return (
        <Row>
            <Col hidden={isApiPasswordOk} md={12}>
                <h5>Zadej heslo:</h5>
                <Form className="d-flex">
                    <Form.Control
                        value={apiPassword}
                        onChange={(e) => changeApiPassword(e.target.value)}
                        type="password"
                        placeholder="Heslo"
                        className="me-2"
                        aria-label="Heslo"
                    />
                </Form>
            </Col>
        </Row>
    )
}

export default PasswordBox;