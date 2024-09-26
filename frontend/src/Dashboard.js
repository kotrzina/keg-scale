import {Container, Row, Toast} from "react-bootstrap";
import {useEffect, useState} from "react";
import "./Dashboard.css";
import Menu from "./Menu";
import Warehouse from "./Warehouse";
import Keg from "./Keg";
import {buildUrl} from "./Api";
import Pivo from "./Pivo";

function Dashboard() {

    const defaultScale = {
        is_ok: false,
        beers_left: 0,
        last_weight: 0.0,
        last_weight_formated: "0.0",
        last_at: "0",
        last_at_duration: "0",
        rssi: 0,
        last_update: 0,
        last_update_duration: 0,
        pub: {
            is_open: false,
            opened_at: 0,
            closed_at: 0,
        },
        active_keg: 0,
    }

    const [scale, setScale] = useState(defaultScale);
    const [showKeg, setShowKeg] = useState(false);
    const [showWarehouse, setShowWarehouse] = useState(false);
    const [showSpinner, setShowSpinner] = useState(false);

    useEffect(() => {
        document.title = "Keg Scale Dashboard"
        void refresh()

        window.addEventListener("focus", refresh)
        const interval = setInterval(() => {
            void refresh()
        }, 10000)

        return () => {
            window.removeEventListener("focus", refresh)
            clearInterval(interval)
        }
        // eslint-disable-next-line
    }, []);

    async function refresh() {
        setShowSpinner(true)
        try {
            const url = buildUrl("/api/scale/dashboard")
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
            <Menu showWarehouse={() => {
                setShowWarehouse(true)
            }} showKeg={() => {
                setShowKeg(true)
            }}/>

            <Warehouse showCanvas={showWarehouse} setShowCanvas={setShowWarehouse}/>
            <Keg keg={scale.active_keg} showCanvas={showKeg} setShowCanvas={setShowKeg} refresh={refresh}/>

            <Row md={12} style={{textAlign: "center", marginTop: "30px"}}>
                <Toast style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            Hospoda&nbsp;&nbsp;
                            <img
                                hidden={!showSpinner}
                                src={"/Rhombus.gif"}
                                width="16"
                                height="16"
                                className="align-middle"
                                alt="Loader"
                            />
                        </strong>
                        <small>{scale.pub.is_open ? scale.pub.opened_at : scale.pub.closed_at}</small>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={scale.pub.is_open ? "cell cell-green" : "cell cell-red"}>
                            {scale.pub.is_open ? "OTEVŘENO" : "ZAVŘENO"}
                        </div>
                    </Toast.Body>
                </Toast>

                <Toast hidden={!scale.is_ok || scale.last_at <= 0} style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            Zbývá piv&nbsp;&nbsp;
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
                            <Pivo amount={scale.beers_left}/>
                        </div>
                    </Toast.Body>
                </Toast>

                <Toast hidden={!scale.pub.is_open || scale.active_keg < 10} style={{margin: "5px"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            Naraženo&nbsp;&nbsp;
                            <img
                                hidden={!showSpinner}
                                src={"/Rhombus.gif"}
                                width="16"
                                height="16"
                                className="align-middle"
                                alt="Loader"
                            />
                        </strong>
                    </Toast.Header>
                    <Toast.Body>
                        <div className={"cell cell-green"}>
                            {scale.active_keg}&nbsp;l
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