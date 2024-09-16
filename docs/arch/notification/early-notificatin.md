# Early Notification

## Why it exists
Using early notification feature, customers are reminded of their reservations earlier than the set time. The value can be changed in configurations.

## How it works
When a customer makes an appointment, following steps will be executed:
- A reservation is created on the database,
- A notification is scheduled to be sent to the customer earlier than the time of appointment.

## Reservation Cancellation
On cancelling an appointment, this notification needs to also be cancelled. So the task scheduler architecture needs to support deleting a scheduled task.

## Technical Decision

API Gateway needs to be blazingly fast. So, task runners should not be executed on the main API Gateway. Task schedulers need to be run on separate instances and the communications between Gateway and schedulers should be established using `Rabbitmq`(worker pattern).

