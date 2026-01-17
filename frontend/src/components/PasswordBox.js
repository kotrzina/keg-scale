import { useAuth } from "../contexts/AuthContext";
import Form from "react-bootstrap/Form";
import { Col, Row } from "react-bootstrap";
import React, { useState } from "react";

function PasswordBox() {
    const { isAuthenticated, login } = useAuth();
    const [inputPassword, setInputPassword] = useState("");

    function handlePasswordChange(e) {
        const newPassword = e.target.value;
        setInputPassword(newPassword);
        login(newPassword);
    }

    return (
        <Row>
            <Col hidden={isAuthenticated} md={12}>
                <h5>Zadej heslo:</h5>
                <Form className="d-flex">
                    <Form.Control
                        value={inputPassword}
                        onChange={handlePasswordChange}
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
