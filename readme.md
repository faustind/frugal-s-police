# AUto BIlling Police

The story

My friend, Paul, has a subscription to the following services: Dazn, Netflix,
Amazon Prime, Spotify, and ExpressVPN. He is on a tight budget for the next
three months since he just moved to Tokyo for work. So he wants to keep his
Dazn subscription only for the hot period of MMA that last one month. But he
forgot to unsubscribe after the months passed, and Dazn took 3000 yen out of
his account. He wasn’t happy because he thought he could use that 3000 yen to
survive one weekend. But instead, he spent them on a service he won’t need.

The solution

I thought I would write a line bot to help Paul keep track of his online
subscriptions and how much they cost him. First, Paul will tell the bot how
long he wants to keep the subscriptions alive. Then, when the lifespan of a
subscription set by Paul is expired, the bot will remind Paul to unsubscribe.

Functionalities

- Users can set a monthly budget for subscriptions.
- Users can specify/edit monthly cost, start date, end date for a subscription
- ~Every month the bot can compute the total amount spent on online subscriptions~
- ~The bot warns for months when the total cost of subscriptions exceeds the budget~.
- The bot informs each subscription when the payday is near


## Deploy

```
./deploy.sh
```

## Messages

- string are case insensitive
- currency is JPY
- format for dates is YYYYMMDD

### ~set budget~

```
yen 'amount'
```

- `yen`: integer 
  - Maximum amount the user plans to spend monthly on online services 

### add subscription

```
eye 'name' 'cost' 'duedate' 'lastmonth'
```

Track due dates (monthly) of subscription

- `name`: string
- `cost`: integer
- `duedate`: YYYYMMDD
- `lastmonth`: YYYMM 
    - The last month the user wishes to pay for the service.
    - the bot will inform you to unsubscribe before the due date in this month

### remove subscription

```
del 'name'
```

Stop tracking the subscription with given name and remove it from db.

- `name`: string
