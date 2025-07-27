const { useState, useEffect } = React;

function CardQueue() {
    const [queue, setQueue] = useState([]);
    const [updated, setUpdated] = useState(new Date());
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const fetchQueue = async () => {
        try {
            const response = await fetch('/cards/queue');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            setQueue(data);
            setError(null);
            setUpdated(new Date());
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        // Initial fetch
        fetchQueue();

        // Set up periodic updates every 5 seconds
        const interval = setInterval(fetchQueue, 5000);

        // Cleanup interval on unmount
        return () => clearInterval(interval);
    }, []);

    const updateStatus = () => {
        const statusElement = document.getElementById('status');
        if (loading) {
            statusElement.textContent = 'Loading...';
        } else if (error) {
            statusElement.textContent = `Error: ${error}`;
        } else {
            statusElement.textContent = `Last updated: ${updated.toLocaleTimeString()}`;
        }
    };

    useEffect(() => {
        updateStatus();
    }, [loading, error, updated]);

    if (loading) {
        return <div className="empty-state">Loading queue...</div>;
    }

    if (error) {
        return <div className="empty-state">Error loading queue: {error}</div>;
    }

    if (queue.length === 0) {
        return <div className="empty-state">No cards in queue</div>;
    }

    return (
        <div>
            {queue.map((jobName, index) => (
                <div key={index} className="queue-item">
                    {jobName}
                </div>
            ))}
        </div>
    );
}

// Render the app
ReactDOM.render(<CardQueue />, document.getElementById('root')); 