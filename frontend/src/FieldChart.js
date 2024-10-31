import {Col, Row, Toast} from "react-bootstrap";
import {Line} from "react-chartjs-2";
import {useEffect, useRef, useState} from "react";
// eslint-disable-next-line
import Chart from 'chart.js/auto';

function FieldChart(props) {

    const chartRef = useRef(null);

    const defaultData = {
        labels: [],
        datasets: [
            {
                label: 'Pivo',
                data: [],
                fill: true,
            },
        ],
    };

    const options = {
        plugins: {
            legend: {
                display: false
            }
        }
    };

    const [activeInterval, setActiveInterval] = useState("1h");
    const [data, setData] = useState(defaultData);

    useEffect(() => {
        if (props.chart.length === 0) {
            return
        }

        const interval = props.chart.find((item) => item.interval === activeInterval);
        if (interval === undefined) {
            return
        }

        setData({
            labels: interval.values.map((item) => item.label),
            datasets: [
                {
                    label: 'Pivo',
                    data: interval.values.map((item) => item.value),
                    fill: true,
                    backgroundColor: 'rgba(69, 57, 32,0.2)',
                    borderColor: 'rgba(219, 166, 55,1)',
                },
            ]
        })
    }, [activeInterval, props.chart]);

    return (
        <Row>
            <Col xs={12} sm={12} md={12} lg={12} xl={12} xxl={12}>
                <Toast style={{width: "100%"}}>
                    <Toast.Header closeButton={false}>
                        <strong className="me-auto">
                            {props.title}&nbsp;&nbsp;
                            <img
                                hidden={!props.loading}
                                src={"/Rhombus.gif"}
                                width="16"
                                height="16"
                                className="align-middle"
                                alt="Loader"
                            />
                        </strong>
                        <small>
                            {props.chart.map((item) => {
                                return (
                                    <span key={item.interval} onClick={(e) => {
                                        e.preventDefault()
                                        setActiveInterval(item.interval)
                                        return false
                                    }} className={activeInterval === item.interval ? "interval activeInterval" : "interval"}>
                                        {item.interval}&nbsp;&nbsp;
                                    </span>
                                )
                            })}
                        </small>
                    </Toast.Header>
                    <Toast.Body>
                        <div>
                            <Line ref={chartRef} data={data} options={options}/>
                        </div>
                    </Toast.Body>
                </Toast>
            </Col>


        </Row>
    )

}

export default FieldChart;