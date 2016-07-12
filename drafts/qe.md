---
hook: Quantitative easing explained in a pragmatic way (and with an example).
image: "/assets/qe/qe2.jpg"
location: San Francisco
published_at: 2014-01-31T06:15:17Z
title: Quantitative Easing for Hackers
attributions: Header image by <strong><a href="http://www.flickr.com/photos/gammaman/6242455757/">Eli
  Christman</a></strong>. Licensed under Creative Commons BY 2.0.
---

As QE3 continues to demonstrate that its nickname of QE-Infinity is well-deserved, it's increasingly important that the general public understands how it works and is able to grasp its long term consequences on the financial system. Is Yellen actually printing $85 billion a month [1]? Does that mean I'll be paying $100 for a loaf of bread by 2020? Looking around the web, I was abhored by just how difficult it was to find an explanation of the precise mechanics of QE; even the Fed itself [assures us that it's not printing](http://www.federalreserve.gov/faqs/money_12853.htm), but then proceeds to gloss over the details.

What follows is my attempt to explain QE in the simple and pragmatic style of the technical blog posts that I've read over the years (thus "for hackers"). Keep in mind that I'm an enthusiast but not an expert, and [welcome feedback and corrections](mailto:brandur@brandur.org).

## Actors

Lets start with the basic and talk about the basic actors involved in the process:

* *United States Department of the Treasury* (or the **Treasury**): the treasury of the US federal government tasked with managing its revenue. The Treasury allows US debt to be purchased by a variety of parties by issuing a financial instrument called a a _US Treasury security_ (these are often known as **treasuries**, and come in a few different forms which we won't get into).
* *Federal Reserve System* (or the **Fed**): the central banking system of the United States. It was created in 1913 as a reaction to an economic panic in 1907 (where the NYSE fell 50% compared to 1906) and tasked by Congress with the responsibilities of maximizing US employment, stabilizing prices, and moderating long-term interest rates. In response to the global financial crisis of 2007-08, it has initiated qualititative easing in an attempt to mitigate the effects of the recession on the US economy. The Fed buys and sells treasuries at auction with its primary dealers.
* *Primary dealers*: ~20 banks permitted to trade directly with the Fed. When the Fed auctions securities, primary dealers are _required_ to participate. The vast majority of treasuries are traded through the primary dealers to other entities worldwide. [Lists of primary dealers](http://en.wikipedia.org/wiki/Primary_dealer#Current_list) are easily available and include well-known names like **BMO**, **Goldman Sachs**, and **JP Morgan**.

## Fractional-reserve Banking

[Fractional-reserve banking](http://en.wikipedia.org/wiki/Fractional_reserve_banking) is a form of banking carried out worldwide that's important to understanding the whole story of QE. Under this system, banks retain reserves only equal to the a _fraction_ of the their total deposits, with that fraction's amount being sufficient to satisfy demand for withdrawals. The rest of the money can be invested or loaned out.

Now here's the catch: that loaned money is often re-deposited into another bank, counts towards that new bank's reserves, and can be re-lent. This magnification effect actually allows the total supply of money to grow to a multiple of what was originally issued by the central bank. This is known as the **[money multiplier](http://en.wikipedia.org/wiki/Money_multiplier)**.

In the US, the Fed stipulates the reserve requirements for US banks and banks operating in US territory, which is the ratio of required reserved to deposits at an institution. For a reasonably large bank (> $79.5M), that ratio is 10%.

The total reserves plus the total currency in circulation is known as the **[monetary base](http://en.wikipedia.org/wiki/Monetary_base)**.

## QE. Step by Step.

Now that we have the backstory in place, let's dive into how QE actually works:

1. The Fed announces that it will buy $X billion worth of treasuries.
1. The primary dealers submit offers to sell those treasuries to the Fed (as they required to do as dictated by their primary dealer status). It's important to note that these primary dealers already own the treasuries that they're going to sell (i.e. they're not created).
1. After the transaction, the primary dealers have now traded the treasuries amongst their assets for a credit that shows up in the books they keep with the Fed.
1. This credit usually becomes part of that bank's _required reserves_ with the Fed, so they can now proceed to lend out more money than they could before. Fractional reserve banking ensures that whatever amount the reserve increased by is effectively multiplied, and therefore its effect on the money supply has a magnified effect, _but_ the increase in reserve isn't cash itself (remember, it's just a credit on the Fed's balance sheet) so the bank has to find liquidity elsewhere if it wants to loan it out.

## Example

Let's throw some numbers into this equation just to illustrate the full effect. We're going to examine the balance sheets of the Treasury, the Fed, and the imaginary primary dealer **Bank A** before and after a round of QE.

<figure>
  <table>
    <tr>
      <th colspan="2">Treasury</th>
      <th colspan="2">Fed</th>
      <th colspan="2">Bank A</th>
    </tr>
    <tr>
      <th>Assets</th>
      <th>Liabilities</th>
      <th>Assets</th>
      <th>Liabilities</th>
      <th>Assets</th>
      <th>Liabilities</th>
    </tr>
    <tr>
      <td>
        <ul>
          <li>$10B public goods</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>$10B treasuries</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>$6B treasuries</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>$6B reserves</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>$1B reserves</li>
          <li>$5B loans</li>
          <li>$4B treasuries</li>
        </ul>
      <td>
        <ul>
          <li>$10B deposits</li>
        </ul>
      </td>
    </tr>
  </table>
  <figcaption>Assets and liabilities of each actor before QE.</figcaption>
</figure>

1. The Fed announces a purchase of $4B worth of treasuries (it should be noted here that where QE1 and QE2 involve the purchase of treasuries, QE1 dealth with [mortgage-backed securities](http://en.wikipedia.org/wiki/Mortgage-backed_securities).
1. In our simplified model, Bank A is the Fed's only primary dealer. It satisfies the Fed's entire demand by selling it $4B of treasuries.
1. $4B appears in Bank A's credit column with the Fed. This number counts toward the required reserve which the Fed has mandated as part of its role of central bank.
1. Although not available as cash, Bank A is now sitting on $4B extra worth of reserve. Before QE, it was just satisfying its reserve requirement by holding $1B for its $10B in deposits. After QE, it can now invest or loan out significantly more money if it deems that the prudent course and can get its hands on the cash to do so.

<figure>
  <table>
    <tr>
      <th colspan="2">Treasury</th>
      <th colspan="2">Fed</th>
      <th colspan="2">Bank A</th>
    </tr>
    <tr>
      <th>Assets</th>
      <th>Liabilities</th>
      <th>Assets</th>
      <th>Liabilities</th>
      <th>Assets</th>
      <th>Liabilities</th>
    </tr>
    <tr>
      <td>
        <ul>
          <li>$10B public goods</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>$10B treasuries</li>
        </ul>
      </td>
      <td>
        <ul>
          <li><strong>$10B treasuries</strong></li>
        </ul>
      </td>
      <td>
        <ul>
          <li><strong>$10B reserves</strong></li>
        </ul>
      </td>
      <td>
        <ul>
          <li><strong>$5B reserves</strong></li>
          <li>$5B loans</li>
        </ul>
      <td>
        <ul>
          <li>$10B deposits</li>
        </ul>
      </td>
    </tr>
  </table>
  <figcaption>Assets and liabilities of each actor after QE.</figcaption>
</figure>

There you have it! I didn't say that the example wouldn't be extremely contrived, but it should serve to illustrate the basic mechanics at work here.

## Is the Fed Printing Money?

Let's circle back around to the original question: is Bernanke printing money and devaluing the dollar? The answer is an unsatisfying "no, but sort of." As we saw above, when the Fed initiates a QE transaction, they're effectively just replacing one type of asset (a treasury) with another type of asset (reserve), neither of which is usable money. However, if from there the bank goes on to use that reserve to increase their outstanding loans, then through the magnification effect of fractional-reserve banking the total money supply _can_ grow, which might edge us ever so slightly closer to that $100 loaf of bread.

That said, the amount of reserve available to a bank is often not the deciding factor when it comes to to extending loans. The Bank for International Settlements (BIS), a global organization for central banks, had this to say about it:

> In fact, the level of reserves hardly figures in banks’ lending decisions. The amount of credit outstanding is determined by banks’ willingness to supply loans, based on perceived risk-return trade-offs, and by the demand for those loans. The aggregate availability of bank reserves does not constrain the expansion directly.

So in other words, the inflationary pressure here may not be that much stronger than it would have otherwise been without QE.

<figure>
  <table>
    <tr>
      <th>Month</th>
      <th>Total monetary base</th>
      <th>Total reserves</th>
      <th>Currency in curculation</th>
    </tr>
    <tr>
      <td>January 2013</td>
      <td>$2,741B</td>
      <td>$1,637B</td>
      <td>$1,159B</td>
    </tr>
    <tr>
      <td>January 2014</td>
      <td>$3,753B</td>
      <td>$2,582B</td>
      <td>$1,227B</td>
    </tr>
    <tr>
      <td>January 2015</td>
      <td>$4,017B</td>
      <td>$2,746B</td>
      <td>$1,333B</td>
    </tr>
  </table>
  <figcaption>Monetary base, reserves, and currency in circulation as published by the Fed.</figcaption>
</figure>

Luckily, we do have some idea as to what's going on as [the Fed tracks the monetary base of the United States](http://www.federalreserve.gov/releases/H3/Current/). The monetary base as of January 2014 is $3,753B, almost a trillion dollars larger than the base of $2,741B in January 2013. That seems like an enormous increase, but if you compare the total reserves of January 2014 ($2,582B) to that of January 2013 ($1,637B) you'll notice that this number is also almost a trillion dollars larger, showing that the increase in reserves is almost entirely responsible for the overall increase to the monetary base. The fact that those reserves are locked up tight &mdash; no bank can lend reserves except to another bank &mdash; means that their effect on inflation is lower than one might expect.

## Other Effects

The Fed's purchase of large numbers of treasuries is enough to push down the **yields** on all government bonds, making them a less attractive investment. This shifts the risk curve and leads investors towards assets like stocks and corporate bonds, pushing the price of those assets higher. These increased prices leads to a **[wealth effect](http://en.wikipedia.org/wiki/Wealth_effect)**, meaning that the increase in perceived wealth improves confidence and leads to more spending, thus stimulating the economy. The effectiveness of this strategy is also difficult to measure directly, but we can observe the price increases in many types of assets over the last few years.

## References

* [Why quantitative easing isn't printing money](http://www.cnbc.com/id/100760150)
* [What is the purpose of QE?](http://www.ritholtz.com/blog/2012/12/what-is-the-purpose-of-qe/)
* [Does Chairman Bernanke Know Squat About Money](http://www.economonitor.com/lrwray/2012/04/19/does-chairman-bernanke-know-squat-about-money/)

[1] When this article was orginally written I'd had Bernanke in place of Yellen as he was still the chairman of the Fed.
