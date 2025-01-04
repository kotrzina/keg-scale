import {Col, Row, Toast} from "react-bootstrap";
import {Line} from "react-chartjs-2";
import {useCallback, useEffect, useMemo, useRef, useState} from "react";
// eslint-disable-next-line
import Chart from 'chart.js/auto';
import FormRange from "react-bootstrap/FormRange";
import {buildUrl} from "./Api";

function FieldChart(props) {

    const chartRef = useRef(null);

    const ranges = useMemo(() => {
        return ["ted", "1h", "2h", "4h", "8h", "12h", "1d", "2d", "3d", "1w", "2w", "1m", "2m", "3m", "6m"]
    }, []);

    const defaultData = useMemo(() => {
        return {
            labels: [],
            datasets: [
                {
                    data: [],
                    fill: true,
                },
            ],
        }
    }, [])

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

    const [activeInterval, setActiveInterval] = useState(0);
    const [data, setData] = useState(defaultData);
    const [loading, setLoading] = useState(true);

    const onIntervalChanged = useCallback(async (interval) => {
        try {
            setLoading(true)
            const range = ranges[interval]
            const url = buildUrl(`/api/scale/chart?metric=${props.metric}&interval=${range}`)
            const res = await fetch(url)
            const response = await res.json()

            setData({
                labels: response.map((item) => item.label),
                datasets: [
                    {
                        data: response.map((item) => item.value),
                        fill: true,
                        backgroundColor: 'rgba(69, 57, 32,0.2)',
                        borderColor: 'rgba(219, 166, 55,1)',
                        stepped: props.stepped,
                        pointRadius: 0,
                    },
                ]
            })
        } catch (e) {
            setData(defaultData)
        } finally {
            setLoading(false)
        }
    }, [defaultData, props.metric, props.stepped, ranges])


    useEffect(() => {
        if (props.defaultRange === undefined) {
            return
        }

        const index = ranges.indexOf(props.defaultRange)
        if (index === -1) {
            return
        }

        setActiveInterval(index)
        void onIntervalChanged(index)
    }, [onIntervalChanged, props.defaultRange, ranges]);

    return (
        <Row className={"mt-3"}>
            <Col xs={12} sm={12} md={12} lg={12} xl={12} xxl={12}>
                <Toast style={{width: "100%"}}>
                    <Toast.Header closeButton={false}>
                        <Row style={{width: "100%", textAlign: "center", margin: 0}}>
                            <Col md={1}>
                                <strong>{ranges[activeInterval]}</strong>
                            </Col>

                            <Col md={9}>
                                <FormRange
                                    min={0}
                                    max={ranges.length - 1}
                                    value={activeInterval}
                                    onChange={e => setActiveInterval(e.target.value)}
                                    onMouseUp={e => onIntervalChanged(e.target.value)}
                                    onTouchEnd={e => onIntervalChanged(e.target.value)}
                                />
                            </Col>

                            <Col md={2}>
                                <strong>{props.title}</strong>&nbsp;&nbsp;
                                <img
                                    hidden={!loading}
                                    src={"/Rhombus.gif"}
                                    width="16"
                                    height="16"
                                    className="align-middle"
                                    alt="Loader"
                                />
                            </Col>
                        </Row>
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