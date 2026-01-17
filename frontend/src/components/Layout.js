import { Outlet } from 'react-router-dom';
import { Container } from 'react-bootstrap';
import Menu from './Menu';
import { DashboardProvider, useDashboard } from '../contexts/DashboardContext';
import Warehouse from './Warehouse';
import Keg from './Keg';
import Bank from './Bank';
import Chat from './Chat';

function LayoutContent() {
    const {
        data,
        refresh,
        showKeg,
        setShowKeg,
        showBank,
        setShowBank,
        showWarehouse,
        setShowWarehouse,
        showChat,
        setShowChat,
    } = useDashboard();

    return (
        <Container>
            <Menu />

            <Warehouse
                warehouse={data.scale.warehouse}
                showCanvas={showWarehouse}
                setShowCanvas={setShowWarehouse}
                refresh={refresh}
            />
            <Keg
                keg={data.scale.active_keg}
                showCanvas={showKeg}
                setShowCanvas={setShowKeg}
                refresh={refresh}
            />
            <Bank
                transactions={data.scale.bank_transactions}
                balance={data.scale.bank_balance}
                showCanvas={showBank}
                setShowCanvas={setShowBank}
                refresh={refresh}
            />
            <Chat
                showCanvas={showChat}
                setShowCanvas={setShowChat}
            />

            <Outlet />
        </Container>
    );
}

function Layout() {
    return (
        <DashboardProvider>
            <LayoutContent />
        </DashboardProvider>
    );
}

export default Layout;
