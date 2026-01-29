import { Alert, Col, Offcanvas, Row, ToastContainer } from "react-bootstrap";
import Form from "react-bootstrap/Form";
import React, { useEffect } from "react";
import { buildUrl } from "../lib/Api";
import Button from "react-bootstrap/Button";
import { useAuth } from "../contexts/AuthContext";
import PasswordBox from "./PasswordBox";

function Chat(props) {

    const { password, isAuthenticated } = useAuth();
    const [showError, setShowError] = React.useState(false)
    const [text, setText] = React.useState("")
    const [messages, setMessages] = React.useState([])

    const [showLoadingMessage, setShowLoadingMessage] = React.useState(false)
    const [loadingText, setLoadingText] = React.useState("Loading")
    let loadingIterator = 0

    useEffect(() => {
        const interval = setInterval(() => {
            loadingIterator++
            const dots = ".".repeat(loadingIterator % 4)
            setLoadingText("Loading" + dots)
        }, 777)

        return () => clearInterval(interval)
    }, [loadingIterator])

    async function send() {
        // add message to messages
        setMessages((curr) => {
            return [{ text: text, from: "me" }, ...curr]
        })

        const request = new Request(buildUrl("/api/ai/chat"), {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": password,
            },
            body: JSON.stringify([{ text: text, from: "me" }, ...messages].reverse()), // reverse to keep order for AI
        });
        setText("")

        setShowLoadingMessage(true)
        const response = await fetch(request)
        if (response.status === 200) {
            setShowError(false)
            const data = await response.json()
            setShowLoadingMessage(false)
            setMessages((curr) => {
                return [{ text: data.text, from: "ai", cost: data.cost }, ...curr]
            })
        } else {
            setShowError(true)
        }

        setShowLoadingMessage(false)
    }

    return (
        <Offcanvas placement={"end"} show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Chat</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>
                <Row hidden={!isAuthenticated}>
                    <Alert hidden={!showError} variant={"danger"}>
                        Chyba! Zkus to prosím později.
                    </Alert>

                    <Form className="d-flex" onSubmit={(e) => {
                        e.preventDefault()
                        void send()
                    }}>
                        <Form.Control
                            size="lg"
                            value={text}
                            onChange={(e) => setText(e.target.value)}
                            type="text"
                            placeholder="Co Tě zajímá?"
                            className="me-2"
                            aria-label="Message"
                        />
                    </Form>

                    <Col md={12} className={"mt-3"}>
                        <Button
                            onClick={() => {
                                void send()
                            }}
                            size={"lg"}
                            variant="success"
                            type="submit"
                            style={{ marginRight: "10px" }}
                        >Odeslat</Button>
                        <Button
                            onClick={() => {
                                setMessages([])
                                setShowError(false)
                            }}
                            hidden={messages.length === 0}
                            size={"lg"}
                            variant="dark"
                            type="submit"
                        >Reset</Button>
                    </Col>

                    <Col md={12} className={"mt-3"}>
                        <ToastContainer className="position-static">
                            <Alert hidden={!showLoadingMessage} key={"default"} className={"mt-2"}
                                variant={"success"}>
                                <Alert.Heading>Pan Botka</Alert.Heading>
                                <p>
                                    {loadingText}
                                </p>
                            </Alert>
                            {messages.map((message, k) => {
                                return (
                                    <Alert key={k} className={"mt-2"}
                                        variant={message.from === "ai" ? "success" : "info"}
                                        title={message.from === "ai" ? `${message.cost.input} / ${message.cost.output}` : ""}
                                    >
                                        <Alert.Heading>{message.from === "ai" ? "Pan Botka" : "Místní štamgast"}</Alert.Heading>
                                        <p dangerouslySetInnerHTML={{ __html: message.text }}></p>
                                    </Alert>
                                )
                            })}
                        </ToastContainer>
                    </Col>
                </Row>

                <PasswordBox />

            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Chat;
