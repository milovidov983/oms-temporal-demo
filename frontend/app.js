let orderId;
async function makeOrder(button) {
    orderId = await createOrder({
        "items": [
            {
                "product_id": "product1",
                "price": 5.0,
                "quantity": 1
            }
        ]
    });
    
    button.textContent = "Заказ создан";
    button.disabled = true;

    const updateButtonContainer = document.getElementById("update-button-container");
    if (!updateButtonContainer.querySelector('.update-btn')) {
        
        const updateButton = document.createElement("button");
        updateButton.textContent = "Обновить";
        updateButton.className = "update-btn";
        updateButton.onclick = function() {
            refreshData(button);
        };
        
        updateButtonContainer.appendChild(updateButton);
    }
}
async function createOrder(body) {
    try {
        const response = await fetch('http://localhost:9999/api/cart', {
            method: 'POST',
            crossDomain: true,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body)
        });
        if (!response.ok) {
            throw new Error('Network response was not ok ' + response.statusText);
        }
        const data = await response.json();
        console.log(data);
        return data.order_id;
    } catch (error) {
        console.error(error);
    }
}
async function refreshData(button) {
    try {
        const response = await fetch('http://localhost:9999/api/cart/status?order_id='+orderId, {
            method: 'GET',
            crossDomain: true
        });
        if (!response.ok) {
            throw new Error('Network response was not ok ' + response.statusText);
        }
        const data = await response.json();
        button.textContent = data.text;
    } catch (error) {
        console.error(error);
    }
}