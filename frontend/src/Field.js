import {Toast} from "react-bootstrap";

function Field(props) {

    function getBodyClass() {
        switch (props.variant) {
            case "green":
                return "cell cell-green";
            case "orange":
                return "cell cell-orange";
            case "red":
                return "cell cell-red";
            default:
                return
        }
    }

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
                <div className={getBodyClass()}>
                    {props.children}
                </div>
            </Toast.Body>
        </Toast>
    )

}

export default Field;