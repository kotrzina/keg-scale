import {Container, Row} from "react-bootstrap";
import {useEffect, useState} from "react";
import "./Dashboard.css";
import Menu from "./Menu";
import Warehouse from "./Warehouse";
import Keg from "./Keg";
import {buildUrl} from "./Api";
import Pivo from "./Pivo";
import Field from "./Field";

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
        is_low: false,
        warehouse: [
            {
                "keg": 10,
                "amount": 0
            },
            {
                "keg": 15,
                "amount": 0
            },
            {
                "keg": 20,
                "amount": 0
            },
            {
                "keg": 30,
                "amount": 0
            },
            {
                "keg": 50,
                "amount": 0
            }
        ]
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

            <Warehouse warehouse={scale.warehouse} showCanvas={showWarehouse} setShowCanvas={setShowWarehouse} refresh={refresh}/>
            <Keg keg={scale.active_keg} showCanvas={showKeg} setShowCanvas={setShowKeg} refresh={refresh}/>

            <Row md={12} style={{textAlign: "center", marginTop: "30px"}}>
                <Field
                    title={"Hospoda"}
                    info={scale.pub.is_open ? scale.pub.opened_at : scale.pub.closed_at}
                    variant={scale.pub.is_open ? "green" : "red"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {scale.pub.is_open ? "OTEVŘENO" : "ZAVŘENO"}
                </Field>

                <Field
                    title={"Zbývá piv"}
                    info={scale.last_at !== "" ? ("před " + scale.last_at_duration) : ""}
                    variant={!scale.pub.is_open ? "red" : scale.is_low ? "orange" : "green"}
                    loading={showSpinner}
                    hidden={false}
                >
                    <Pivo amount={scale.beers_left}/>
                </Field>

                <Field
                    title={"Bečka"}
                    variant={!scale.pub.is_open ? "red" : "green"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {scale.active_keg}&nbsp;l
                </Field>

                <Field
                    title={"Váha"}
                    info={"před " + scale.last_at_duration}
                    variant={"green"}
                    loading={showSpinner}
                    hidden={!scale.is_ok || scale.last_at <= 0}
                >
                    {scale.last_weight_formated}&nbsp;kg
                </Field>

                <Field
                    title={"Sklad"}
                    info={""}
                    variant={scale.pub.is_open ? "green" : "red"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {scale.warehouse.reduce((acc, keg) => acc + keg.amount + " ", "")}
                </Field>

                <Field
                    title={"Status"}
                    info={"před " + scale.last_update_duration}
                    variant={scale.pub.is_open ? "green" : "red"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {scale.is_ok ? "OK" : "OFFLINE"}
                </Field>

                <Field
                    title={"WiFi"}
                    info={"před " + scale.last_update_duration}
                    variant={"green"}
                    loading={showSpinner}
                    hidden={!scale.is_ok}
                >
                    {scale.rssi}&nbsp;db
                </Field>

            </Row>
        </Container>
    )
}

export default Dashboard;