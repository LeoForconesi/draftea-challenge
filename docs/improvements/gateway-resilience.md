# Gateway Resilience Improvements
↩️ [Return to README](../../README.md)

A key improvement to enhance the resilience of the Wallet API service is the implementation of robust error handling 
and retry mechanisms when interacting with the external payment gateway. 
This ensures that transient failures do not lead to lost transactions or inconsistent states within the wallet system.

We have a couple of possible solutions to improve gateway resilience:
1. Async Outbox + retry queue
Instead of blocking the payment request on gateway availability, we can decouple the payment processing using an outbox pattern combined with a retry queue.
The flow would be:
- On payment request, create a PENDING payment transaction and an outbox event.
- A new worker (to  be created) processes the outbox event, calling the payment gateway.
- We could use a DLQ (Dead Letter Queue) for failed payments after several retries.
- On successful payment, update the transaction status to APPROVED/REJECTED and publish an event. 
The only thing we must consider is thath the payment status becomes eventually consistent, as the client will not receive immediate confirmation of success/failure.

2. DQL with manual process or batch
This is useful when the gateway fails for hours/days. You can move "failed" records to a DLQ table/queue and retry later.
Operationally, it usually requires metrics/alerts. 
3. What I mean by "manual" is provider some endpoint or admin UI to reprocess failed payments in the DLQ, or have a scheduled batch job that retries them periodically.
