import { NavLink } from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import Nav from 'react-bootstrap/Nav';
import Navbar from 'react-bootstrap/Navbar';
import { useDashboard } from '../contexts/DashboardContext';

function Menu() {
    const { setShowKeg, setShowWarehouse, setShowChat, setShowBank } = useDashboard();

    return (
        <Navbar expand="lg" className="bg-body-tertiary">
            <Container fluid>
                <Navbar.Brand as={NavLink} to="/">Pub</Navbar.Brand>
                <Navbar.Toggle aria-controls="navbarScroll" />
                <Navbar.Collapse id="navbarScroll">
                    <Nav
                        className="me-auto my-2 my-lg-0"
                        style={{ maxHeight: '150px' }}
                        navbarScroll
                    >
                        <Nav.Link as={NavLink} to="/" end>
                            Dashboard
                        </Nav.Link>

                        <Nav.Link onClick={() => setShowKeg(true)}>
                            Keg
                        </Nav.Link>

                        <Nav.Link onClick={() => setShowWarehouse(true)}>
                            Sklad
                        </Nav.Link>

                        <Nav.Link onClick={() => setShowChat(true)}>
                            Chat
                        </Nav.Link>

                        <Nav.Link onClick={() => setShowBank(true)}>
                            Banka
                        </Nav.Link>
                    </Nav>
                </Navbar.Collapse>

                <Navbar.Collapse className="justify-content-end">
                    <a
                        href="https://github.com/kotrzina/keg-scale"
                        target="_blank"
                        rel="noreferrer"
                    >
                        <img src="/github-mark.png" width="32px" alt="Github.com logo" />
                    </a>
                </Navbar.Collapse>
            </Container>
        </Navbar>
    );
}

export default Menu;
