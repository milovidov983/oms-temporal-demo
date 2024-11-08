POST http://localhost:8888/api/orders
Content-Type: application/json
{
    "customer_id": "customer456",
    "items": [
        {
            "product_id": "product789",
            "quantity": 2,
            "price": 150.0
        },
        {
            "product_id": "product101",
            "quantity": 1,
            "price": 200.0
        }
    ],
    "total_amount": 500.0,
    "status": "NEW",
    "created_at": "2024-11-05T12:00:00Z"
}

HTTP/1.1 200
[Asserts]
jsonpath "$.order_id" exists
