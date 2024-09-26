function Pivo(props) {

    if (props.amount === 0) {
        return (
            <>
                Prázdné
            </>
        )
    }

    if (props.amount === 1) {
        return (
            <>
                {props.amount}&nbsp;pivo
            </>
        )
    }

    if (props.amount > 0 && props.amount < 5) {
        return (
            <>
                {props.amount}&nbsp;piva
            </>
        )
    }

    return (
        <>
            {props.amount}&nbsp;piv
        </>
    )
}

export default Pivo;