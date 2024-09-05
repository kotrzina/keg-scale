import {Container, Row, Toast} from "react-bootstrap";
import {useEffect, useState} from "react";
import "./Dashboard.css";
import Menu from "./Menu";
import Warehouse from "./Warehouse";

function Dashboard() {

    const defaultScale = {
        is_ok: false,
        last_weight: 0.0,
        last_weight_formated: "0.0",
        last_at: "0",
        last_at_duration: "0",
        rssi: 0,
        last_update: 0,
        last_update_duration: 0,
    }

    const [scale, setScale] = useState(defaultScale);
    const [showCanvas, setShowCanvas] = useState(false);
    const [showSpinner, setShowSpinner] = useState(false);

    useEffect(() => {
        document.title = "Keg Scale Dashboard"
        void update()

        window.addEventListener("focus", update)
        const interval = setInterval(() => {
            void update()
        }, 10000)

        return () => {
            window.removeEventListener("focus", update)
            clearInterval(interval)
        }
        // eslint-disable-next-line
    }, []);

    async function update() {
        setShowSpinner(true)
        try {
            // REACT_APP_BACKEND_PREFIX is defined in .env file for development
            // and it is empty for production because the backend is on the same domain and port
            let url = "/api/scale/dashboard"
            if (process.env.REACT_APP_BACKEND_PREFIX !== undefined) {
                url = process.env.REACT_APP_BACKEND_PREFIX + "/api/scale/dashboard"
            }

            const res = await fetch(url)
            if (res.statusCode === 425) {
                setScale(defaultScale)
                setShowSpinner(false)
                return // scale does not have any data yet
            }

            const data = await res.json()
            setScale(data)
        } catch {
            setScale(defaultScale)
        }
        setShowSpinner(false)
    }

    return (
        <Container>
            <Menu showCanvas={() => {
                setShowCanvas(true)
            }}/>

            <Warehouse showCanvas={showCanvas} setShowCanvas={setShowCanvas}/>

            <Row md={12} style={{textAlign: "center", marginTop: "30px"}}>
                <Toast style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            Status&nbsp;&nbsp;
                            <img
                                hidden={!showSpinner}
                                src={"/Rhombus.gif"}
                                width="16"
                                height="16"
                                className="align-middle"
                                alt="Loader"
                            />
                        </strong>
                        <small>před {scale.last_update_duration}</small>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={scale.is_ok ? "cell cell-green" : "cell cell-red"}>
                            {scale.is_ok ? "OK" : "OFFLINE"}
                        </div>
                    </Toast.Body>
                </Toast>

                <Toast hidden={!scale.is_ok || scale.last_at <= 0} style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            Váha&nbsp;&nbsp;
                            <img
                                hidden={!showSpinner}
                                src={"/Rhombus.gif"}
                                width="16"
                                height="16"
                                className="align-middle"
                                alt="Loader"
                            />
                        </strong>
                        <small>před {scale.last_at_duration}</small>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={"cell cell-green"}>
                            {scale.last_weight_formated} kg
                        </div>
                    </Toast.Body>
                </Toast>

                <Toast hidden={!scale.is_ok} style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            WiFi&nbsp;&nbsp;
                            <img
                                hidden={!showSpinner}
                                src={"/Rhombus.gif"}
                                width="16"
                                height="16"
                                className="align-middle"
                                alt="Loader"
                            />
                        </strong>
                        <small>před {scale.last_update_duration}</small>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={"cell cell-green"}>
                            {scale.rssi} db
                        </div>
                    </Toast.Body>
                </Toast>

            </Row>
        </Container>
    )
}

export default Dashboard;