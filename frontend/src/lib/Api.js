// REACT_APP_BACKEND_PREFIX is defined in .env file for development
// and it is empty for production because the backend is on the same domain and port
export function buildUrl(endpoint) {
    let url = endpoint
    if (process.env.REACT_APP_BACKEND_PREFIX !== undefined) {
        url = process.env.REACT_APP_BACKEND_PREFIX + endpoint
    }

    return url
}
