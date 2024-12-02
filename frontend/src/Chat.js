import {Alert, Col, Offcanvas, Row, ToastContainer} from "react-bootstrap";
import Form from "react-bootstrap/Form";
import React, {useEffect} from "react";
import {buildUrl} from "./Api";
import Button from "react-bootstrap/Button";

function Chat(props) {

    const [passwordHidden, setPasswordHidden] = React.useState(false)
    const [showError, setShowError] = React.useState(false)
    const [text, setText] = React.useState("")
    const [password, setPassword] = React.useState("")
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

    function htmlDecode(content) {
        let e = document.createElement('div');
        e.innerHTML = content;
        return e.innerHTML
    }

    function urlify(text) {
        const urlRegex = /(https?:\/\/[^\s]+)/g;
        return text.replace(urlRegex, function (url) {
            return htmlDecode('<a target="_blank" href="' + url + '">' + url + '</a>');
        })
    }

    async function send() {
        // add message to messages
        setMessages((curr) => {
            return [{text: text, from: "me"}, ...curr]
        })

        const request = new Request(buildUrl("/api/ai/test"), {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": password,
            },
            body: JSON.stringify([{text: text, from: "me"}, ...messages].reverse()), // reverse to keep order for AI
        });
        setText("")

        setShowLoadingMessage(true)
        const response = await fetch(request)
        if (response.status === 200) {
            setPasswordHidden(true)
            setShowError(false)
            const body = await response.text()
            setShowLoadingMessage(false)
            setMessages((curr) => {
                return [{text: body, from: "ai"}, ...curr]
            })
        } else {
            setShowError(true)
        }

        setShowLoadingMessage(false)
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
        <Offcanvas placement={"end"} show={props.showCanvas} onHide={() => {
            props.setShowCanvas(false)
        }}>
            <Offcanvas.Header closeButton>
                <Offcanvas.Title>Chat</Offcanvas.Title>
            </Offcanvas.Header>
            <Offcanvas.Body>
                <Row>

                    <Alert hidden={!showError} variant={"danger"}>
                        Chyba! Asi špatné heslo.
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
                            style={{marginRight: "10px"}}
                        >Odeslat</Button>
                        <Button
                            onClick={() => {
                                setMessages([])
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
                                           variant={message.from === "ai" ? "success" : "info"}>
                                        <Alert.Heading>{message.from === "ai" ? "Pan Botka" : "Místní štamgast"}</Alert.Heading>
                                        <p dangerouslySetInnerHTML={{__html: urlify(message.text)}}></p>

                                    </Alert>
                                )
                            })}
                        </ToastContainer>
                    </Col>

                    <Col hidden={passwordHidden} md={12} className={"mt-5"}>
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
                    </Col>
                </Row>


            </Offcanvas.Body>
        </Offcanvas>
    )
}

export default Chat;