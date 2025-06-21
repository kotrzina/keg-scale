import {Col, Offcanvas, Row, Table} from "react-bootstrap";
import React from "react";

function Bank(props) {
    return (
        <Offcanvas show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Poslední trasakce</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>
                <Row>
                    <Col md={12}>
                        <Table>
                            <thead>
                            <tr>
                                <th>Datum</th>
                                <th>Popis</th>
                                <th>Částka</th>
                            </tr>
                            </thead>
                            <tbody>
                            {props.transactions.slice().reverse().map((transaction, index) => (
                                <tr key={index}>
                                    <td>
                                        {transaction.date
                                            ? new Date(transaction.date).toLocaleDateString("cs-CZ", {
                                                day: "numeric",
                                                month: "numeric"
                                            })
                                            : ""}
                                    </td>
                                    <td>{transaction.account_name}</td>
                                    <td className={transaction.amount > 0 ? "text-success" : "text-danger"}>
                                        {transaction.amount} Kč
                                    </td>

                                </tr>
                            ))}
                            </tbody>
                        </Table>
                    </Col>
                </Row>

            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Bank;