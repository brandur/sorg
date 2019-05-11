+++
hook = "UNWRITTEN. This should not appear on the front page."
published_at = 2016-07-20T04:48:24Z
title = "us-east-1"
+++

I have fond memories of operating Heroku in us-east-1. Be it the electrical
storm of 2012, or the ELB failures of Christmas Eve, the region is plagued by
more than its fair share of outages and problems. With Heroku being entirely
single region, these major cross-AZ issues would take us down every time. Here
are a few favorites and links to their postmorterms:

* [Apr 20, 2011](https://aws.amazon.com/message/65648/): Stuck EBS volumes
  lasting two days.
* [Jun 29, 2012](https://aws.amazon.com/message/67457/): Electrical storm in
  North Virginia and backup generator problems.
* [Oct 22, 2012](https://aws.amazon.com/message/680342/): Stuck EBS volumes.
* [Dec 24, 2012](https://aws.amazon.com/message/680587/): ELB failures.
* [Sep 20, 2015](https://aws.amazon.com/message/5467D2/): DynamoDB outage and
  cascading failure.

Today, us-east-1 experienced yet another major API outage; with the API being
degraded or completely down completely for two hours between 1115a and 115p PT.
True to their reputation, Amazon was slow to update their status page, leaving
users struggling to understand what was going on for themselves.

About a year ago, Stripe undertook the major project of moving all its servers
over to us-west-2 (Oregon), and completed it within the last year. The project
had a few major objectives, getting our fleet onto VPC being one for example,
but among them was the simple goal of anchoring into one of AWS' most modern
and problem-free data centers.

The project involved an incredible amount of effort on the part of many people,
but seems to be paying off already. While my friends at places like Heroku and
Citus labored under duress waiting for AWS to come back online, the silence
over on the west coast was deafening.
