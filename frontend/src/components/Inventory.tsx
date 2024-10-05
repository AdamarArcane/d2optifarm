import React, { useState } from 'react';
import { Button } from './ui/button'

// Define types for your inventory items
interface InventoryItem {
    itemHash: number;
    itemInstanceId: string;
    quantity: number;
    // Include other fields if available
}

const Inventory: React.FC = () => {
    const [inventory, setInventory] = useState<InventoryItem[] | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetchInventory = async () => {
        try {
            const response = await fetch('/api/inventory', {
                method: 'GET',
                credentials: 'include', // Include cookies for authentication
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (response.ok) {
                const data: InventoryItem[] = await response.json();
                setInventory(data);
                setError(null);
            } else {
                setError('Failed to fetch inventory');
            }
        } catch (err) {
            setError('An error occurred while fetching inventory');
        }
    };

    return (
        <div className="inventory-container">
            <h2>Your Inventory</h2>
            <Button onClick={fetchInventory}>Fetch Inventory</Button>

            {error && <p style={{ color: 'red' }}>{error}</p>}

            {inventory && (
                <div className="inventory-list">
                    {inventory.map((item) => (
                        <div key={item.itemInstanceId} className="inventory-item">
                            <p>Item Hash: {item.itemHash}</p>
                            <p>Quantity: {item.quantity}</p>
                            {/* Display more item details if available */}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default Inventory;