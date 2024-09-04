import {render, screen} from '@testing-library/react';
import Dashboard from "./Dashboard";


test('renders learn react link', () => {
    render(<Dashboard/>);
    const scaleText = screen.getByText(/Váha/i);
    expect(scaleText).toBeInTheDocument();

    const wifiText = screen.getByText(/WiFi/i);
    expect(wifiText).toBeInTheDocument();
});
