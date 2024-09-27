import {Toast} from "react-bootstrap";

function Field(props) {

    return (
        <Toast style={{margin: "5px"}} hidden={props.hidden}>
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
                <small>{props.info}</small>
            </Toast.Header>
            <Toast.Body>
                <div className={props.variant === "green" ? "cell cell-green" : "cell cell-red"}>
                    {props.children}
                </div>
            </Toast.Body>
        </Toast>
    )

}

export default Field;