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

        // Set up periodic updates every 3 seconds
        const interval = setInterval(fetchQueue, 3000);

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

    const getJobWrapperClassName = (state) => {
        switch (state) {
            case 'queue':
                return 'queue-item job-queue-wrapper';
            case 'process':
                return 'queue-item job-process-wrapper';
            case 'done':
                return 'queue-item job-done-wrapper';
            case 'failed':
                return 'queue-item job-failed-wrapper';
            default:
                return 'queue-item job-queue-wrapper';
        }
    };

    const getJobClassName = (state) => {
        switch (state) {
            case 'queue':
                return 'job-queue';
            case 'process':
                return 'job-process';
            case 'done':
                return 'job-done';
            case 'failed':
                return 'job-failed';
            default:
                return 'job-queue';
        }
    };


    const getAnomationClassName = (state) => {
        switch (state) {
            case 'queue':
                return 'job-queue-animation';
            case 'process':
                return 'job-process-animation';
            case 'done':
                return 'job-done-animation';
            case 'failed':
                return 'job-failed-animation';
            default:
                return 'job-queue-animation';
        }
    };


    return (
        <div>
            {queue.map((job, index) => {
                const maxLength = 16;
                return (
                    <div key={index} className={getJobWrapperClassName(job.state)}>
                        <div className={getAnomationClassName(job.state)}>
                            <div className={getJobClassName(job.state)}>
                                {job.name.length > maxLength ? job.name.substring(0, maxLength) + '...' : job.name}
                            </div>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}

// Render the app
ReactDOM.render(<CardQueue />, document.getElementById('root')); 