import {Container, Row, Toast} from "react-bootstrap";
import {useEffect, useState} from "react";
import "./Dashboard.css";


function Dashboard() {

    const [weight, setWeight] = useState(0);
    const [lastUpdate, setLastUpdate] = useState("");

    const [rssi, setRssi] = useState(1);

    useEffect(() => {
        update()
        const interval = setInterval(() => {
            update()
        }, 5000)
        return () => clearInterval(interval)
    }, []);

    async function update() {
        try {
            const data = await fetch("http://localhost:8080/api/scale/dashboard")
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