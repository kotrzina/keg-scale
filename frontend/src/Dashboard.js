import {Container, Row} from "react-bootstrap";
import {useEffect, useState} from "react";
import "./Dashboard.css";
import Menu from "./Menu";
import Warehouse from "./Warehouse";
import Keg from "./Keg";
import {buildUrl} from "./Api";
import Pivo from "./Pivo";
import Field from "./Field";
import FieldChart from "./FieldChart";

function Dashboard() {

    const defaultEmptyChart = [{
        interval: "1h",
        values: [
            {
                label: "0",
                value: 0
            }
        ]
    }];

    const defaultScale = {
        scale: {
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
            ],
            warehouse_beer_left: 0,
        },

        charts: {
            beers_left: defaultEmptyChart,
        }
    }


    const [data, setData] = useState(defaultScale);
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
            setData(data)
        } catch {
            setData(defaultScale)
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

            <Warehouse warehouse={data.scale.warehouse} showCanvas={showWarehouse} setShowCanvas={setShowWarehouse}
                       refresh={refresh}/>
            <Keg keg={data.scale.active_keg} showCanvas={showKeg} setShowCanvas={setShowKeg} refresh={refresh}/>

            <Row md={12} style={{textAlign: "center", marginTop: "30px"}}>
                <Field
                    title={"Hospoda"}
                    info={data.scale.pub.is_open ? data.scale.pub.opened_at : data.scale.pub.closed_at}
                    variant={data.scale.pub.is_open ? "green" : "red"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {data.scale.pub.is_open ? "OTEVŘENO" : "ZAVŘENO"}
                </Field>

                <Field
                    title={"Zbývá piv"}
                    info={data.scale.last_at !== "" ? ("před " + data.scale.last_at_duration) : ""}
                    variant={!data.scale.pub.is_open ? "red" : data.scale.is_low ? "orange" : "green"}
                    loading={showSpinner}
                    hidden={false}
                >
                    <Pivo amount={data.scale.beers_left}/>
                </Field>

                <Field
                    title={"Bečka"}
                    variant={!data.scale.pub.is_open ? "red" : "green"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {data.scale.active_keg}&nbsp;l
                </Field>

                <Field
                    title={"Váha"}
                    info={"před " + data.scale.last_at_duration}
                    variant={"green"}
                    loading={showSpinner}
                    hidden={!data.scale.is_ok || data.scale.last_at <= 0}
                >
                    {data.scale.last_weight_formated}&nbsp;kg
                </Field>

                <Field
                    title={"Sklad"}
                    info={""}
                    variant={"green"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {data.scale.warehouse.reduce((acc, keg) => acc + keg.amount + " ", "")}
                </Field>

                <Field
                    title={"Sklad"}
                    info={""}
                    variant={data.scale.warehouse_beer_left > 100 ? "green" : "orange"}
                    loading={showSpinner}
                    hidden={data.scale.warehouse_beer_left <= 0}
                >
                    {data.scale.warehouse_beer_left}&nbsp;piv
                </Field>

                <Field
                    title={"Status"}
                    info={"před " + data.scale.last_update_duration}
                    variant={data.scale.pub.is_open ? "green" : "red"}
                    loading={showSpinner}
                    hidden={false}
                >
                    {data.scale.is_ok ? "OK" : "OFFLINE"}
                </Field>

                <Field
                    title={"WiFi"}
                    info={"před " + data.scale.last_update_duration}
                    variant={"green"}
                    loading={showSpinner}
                    hidden={!data.scale.is_ok}
                >
                    {data.scale.rssi}&nbsp;db
                </Field>
            </Row>

            <FieldChart title={"Zbýva piva"} chart={data.charts.beers_left} loading={showSpinner}/>


            <Row className={"mt-4"}></Row>
        </Container>
    )
}

export default Dashboard;