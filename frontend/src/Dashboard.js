import {Container, Row, Toast} from "react-bootstrap";
import {useEffect, useState} from "react";
import "./Dashboard.css";

function Dashboard() {

    const [weight, setWeight] = useState(0);
    const [lastUpdate, setLastUpdate] = useState("");

    const [rssi, setRssi] = useState(1);

    useEffect(() => {
        void update()
        const interval = setInterval(() => {
            void update()
        }, 10000)
        return () => clearInterval(interval)
    }, []);

    async function update() {
        try {
            // REACT_APP_BACKEND_PREFIX is defined in .env file for development
            // and it is empty for production because the backend is on the same domain and port
            const data = await fetch(process.env.REACT_APP_BACKEND_PREFIX + "/api/scale/dashboard")
            const json = await data.json()

            setWeight(json.weight_formated)
            setLastUpdate(json.last_update_duration)
            setRssi(json.rssi)
        } catch {
            setLastUpdate("error")
            setWeight("error")
            setRssi("")
        }
    }

    return (
        <Container>
            <Row md={12} style={{textAlign: "center", marginTop: "30px"}}>
                <Toast style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">Váha</strong>
                        <small>{lastUpdate} ago</small>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={"cell"}>
                            {weight}<span hidden={weight === "error"}> kg</span>
                        </div>
                    </Toast.Body>
                </Toast>
                <Toast style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">WiFi</strong>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={"cell"}>
                            {rssi}<span hidden={rssi === ""}> db</span>
                        </div>
                    </Toast.Body>
                </Toast>


            </Row>
        </Container>
    )
}

export default Dashboard;