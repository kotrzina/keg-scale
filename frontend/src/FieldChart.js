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
                data: [],
                fill: true,
            },
        ],
    };

    const options = {
        scales: {
            y: {
                beginAtZero: true
            }
        },
        plugins: {
            legend: {
                display: false
            }
        }
    };

    const DEFAULT_INTERVAL = "NO_DATA";

    const [activeInterval, setActiveInterval] = useState(DEFAULT_INTERVAL);
    const [data, setData] = useState(defaultData);

    useEffect(() => {
        if (props.chart === undefined) {
            return
        }

        if (props.chart.length <= 0) {
            return
        }

        let i = activeInterval
        if (i === DEFAULT_INTERVAL) {
            // find first interval with data from the end
            for (let j = props.chart.length - 1; j >= 0; j--) {
                if (props.chart[j].values !== null && props.chart[j].values.length > 0) {
                    i = props.chart[j].interval
                    break
                }
            }
            setActiveInterval(i)
            return
        }

        const interval = props.chart.find((item) => item.interval === i);
        if (interval === undefined) {
            return
        }

        if (interval.values === null) {
            return
        }

        setData({
            labels: interval.values.map((item) => item.label),
            datasets: [
                {
                    data: interval.values.map((item) => item.value),
                    fill: true,
                    backgroundColor: 'rgba(69, 57, 32,0.2)',
                    borderColor: 'rgba(219, 166, 55,1)',
                    stepped: props.stepped,
                    pointRadius: 0,
                },
            ]
        })
    }, [activeInterval, props.chart, props.stepped]);

    return (
        <Row className={"mt-3"}>
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
                                    <span hidden={item.values === null} key={item.interval} onClick={(e) => {
                                        e.preventDefault()
                                        setActiveInterval(item.interval)
                                        return false
                                    }}
                                          className={activeInterval === item.interval ? "interval activeInterval" : "interval"}>
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