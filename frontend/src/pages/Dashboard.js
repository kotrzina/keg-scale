import { Row } from "react-bootstrap";
import { useEffect } from "react";
import "./Dashboard.css";
import Pivo from "../components/Pivo";
import Field from "../components/Field";
import FieldChart from "../components/FieldChart";
import { useDashboard } from "../contexts/DashboardContext";

function Dashboard() {
    const { data, isLoading } = useDashboard();

    useEffect(() => {
        document.title = "Keg Scale Dashboard";
    }, []);

    return (
        <>
            <Row md={12} style={{ textAlign: "center", marginTop: "30px" }}>
                <Field
                    title={"Hospoda"}
                    info={data.scale.pub.is_open ? data.scale.pub.opened_at : data.scale.pub.closed_at}
                    variant={data.scale.pub.is_open ? "green" : "red"}
                    loading={isLoading}
                    hidden={false}
                >
                    {data.scale.pub.is_open ? "OTEVŘENO" : "ZAVŘENO"}
                </Field>

                <Field
                    title={"Zbývá piv"}
                    info={data.scale.last_at !== "" ? ("před " + data.scale.last_at_duration) : ""}
                    variant={!data.scale.pub.is_open ? "red" : data.scale.is_low ? "orange" : "green"}
                    loading={isLoading}
                    hidden={false}
                >
                    <Pivo amount={data.scale.beers_left} />
                </Field>

                <Field
                    title={"Bečka"}
                    variant={!data.scale.pub.is_open ? "red" : "green"}
                    loading={isLoading}
                    hidden={false}
                >
                    {data.scale.active_keg}&nbsp;l
                </Field>

                <Field
                    title={"Váha"}
                    info={"před " + data.scale.last_at_duration}
                    variant={"green"}
                    loading={isLoading}
                    hidden={!data.scale.is_ok || data.scale.last_at <= 0}
                >
                    {data.scale.last_weight_formated}&nbsp;kg
                </Field>

                <Field
                    title={"Celkem piv"}
                    info={"od 12.11.2024"}
                    variant={"green"}
                    loading={isLoading}
                    hidden={false}
                >
                    {data.scale.beers_total}&nbsp;piv
                </Field>

                <Field
                    title={"Sklad"}
                    info={""}
                    variant={"green"}
                    loading={isLoading}
                    hidden={false}
                >
                    {data.scale.warehouse.reduce((acc, keg) => acc + keg.amount + " ", "")}
                </Field>

                <Field
                    title={"Sklad"}
                    info={""}
                    variant={data.scale.warehouse_beer_left > 100 ? "green" : "orange"}
                    loading={isLoading}
                    hidden={data.scale.warehouse_beer_left <= 0}
                >
                    {data.scale.warehouse_beer_left}&nbsp;piv
                </Field>

                <Field
                    title={"Status"}
                    info={"před " + data.scale.last_update_duration}
                    variant={data.scale.pub.is_open ? "green" : "red"}
                    loading={isLoading}
                    hidden={false}
                >
                    {data.scale.is_ok ? "OK" : "OFFLINE"}
                </Field>

                <Field
                    title={"WiFi"}
                    info={"před " + data.scale.last_update_duration}
                    variant={"green"}
                    loading={isLoading}
                    hidden={!data.scale.is_ok}
                >
                    {data.scale.rssi}&nbsp;db
                </Field>

                <Field
                    title={"Banka"}
                    info={"před " + data.scale.last_update_duration}
                    variant={"green"}
                    loading={isLoading}
                    hidden={!data.scale.bank_balance.balance}
                >
                    {data.scale.bank_balance.balance}&nbsp;CZK
                </Field>
            </Row>

            <FieldChart title={"Zbýva piva"} metric={"scale_beers_left"} defaultRange="ted" stepped={false} />
            <FieldChart title={"Aktivní bečka"} metric={"scale_active_keg"} defaultRange="2w" stepped={true} />

            <Row className={"mt-4"}></Row>
        </>
    );
}

export default Dashboard;
