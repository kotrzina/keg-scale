import Container from 'react-bootstrap/Container';
import Nav from 'react-bootstrap/Nav';
import Navbar from 'react-bootstrap/Navbar';

function Menu(props) {
    return (
        <Navbar expand="lg" className="bg-body-tertiary">
            <Container fluid>
                <Navbar.Brand href="#">Pub</Navbar.Brand>
                <Navbar.Toggle aria-controls="navbarScroll"/>
                <Navbar.Collapse id="navbarScroll">
                    <Nav
                        className="me-auto my-2 my-lg-0"
                        style={{maxHeight: '150px'}}
                        navbarScroll
                    >

                        <Nav.Link onClick={() => {
                            props.showKeg()
                        }}>Keg</Nav.Link>

                        <Nav.Link onClick={() => {
                            props.showWarehouse()
                        }}>Sklad</Nav.Link>

                        <Nav.Link onClick={() => {
                            props.showChat()
                        }}>Chat</Nav.Link>

                    </Nav>
                </Navbar.Collapse>

                <Navbar.Collapse className="justify-content-end">
                        <a
                            href={"https://github.com/kotrzina/keg-scale"}
                            target={"_blank"}
                            rel={"noreferrer"}
                        >
                            <img src={"/github-mark.png"} width={"32px"} alt={"Github.com logo"}/>
                        </a>
                </Navbar.Collapse>
            </Container>
        </Navbar>
    );
}

export default Menu;