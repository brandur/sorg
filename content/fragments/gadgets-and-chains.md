---
title: "Depickling, gadgets, and chains: The class of
  exploit that unraveled Equifax"
published_at: 2017-09-10T19:42:06Z
hook: How unsafe deserialization leads to arbitrary code
  execution.
---

Anyone with an internet connection is probably aware that
Equifax recently leaked information on 143 million people
in America; resulting in one of the most impactful security
breaches in history.

The problem was reported to be in [Apache Struts][struts],
a Rails-esque MVC framework for the Java ecosystem. A
likely candidate [1] is [CVE-2017-9805][strutsvuln], which
allowed an attacker to send a malicious XML payload that
would be hydrated by XStream (an XML serialization library)
to an arbitrary object.

***Update:*** Contrary to what I've said above, it's been
made public that the vulnerability was
[CVE-2017-5638][strutsvuln2] which allowed arbitrary
command execution through a maliciously constructed
`Content-Type` header. Unlike CVE-2017-9805, its details
_are_ disclosed. See here the [Metasploit pull
request][metasploit] that adds a proof of concept.

## Unsafe depickling (#unsafe-depickling)

The exact details of the exploit are still unknown, but the
broad class of exploit is well-articulated in the [2015
talk _Marshalling Pickles_][talk] by Gabriel Lawrence and
Chris Frohoff.

Many languages provide a mechanism to allow objects to be
encoded or decoded for transport. In Java, C#, and PHP it's
called "serialization"; in Ruby "marshalling"; in Go
"gobbing"; and in Python "pickling". Each of an object's
fields are written to a binary or string-based format along
with an identifier for its class. A decoder running the
same code on the other size instantiates the class and sets
each field to stored values to reconstruct the serialized
object. This sounds really scary and you might ask why
these packages event exist, but they turn out to be very
useful when implementing features like RPC.

It's the combination of these generic deserialization
mechanisms combined with unchecked input that's big enough
to lead to the leak of PII for 143 million people.

## Gadgets and chains (#gadgets-and-chains)

Lawrence & Frohoff use the term ***gadget*** to describe a
class or function that's available within in executing
scope of an application. The exploitation strategy is to
start with a "kick-off" gadget that's executed after
deserialization and build a ***chain*** of instances and
method invocations to get to a "sink" gadget that's able to
execute arbitrary code or commands. After attackers manage
to get input to a sink gadget, they've effectively found a
way to own the box.

!fig src="/assets/images/fragments/gadgets-and-chains/gadgets-and-chains.svg" caption="Child processes transitioning from mostly shared memory to mostly copied as they mature."

Here's a basic example of a Java gadget chain in action
(lifted more-or-less unchanged from [the talk][talk]):

``` java
public class CacheManager implements Serializable {
    private final Runnable initHook;

    public void readObject(ObjectInputStream ois) {
        ois.defaultReadObject(); // Populate initHook
        initHook.run();
    }
}

public class CommandTask implements Runnable, Serializable {
    private final String command;

    // For example, "cmd.exe" is passed into command.
    public CommandTask(String command) {
        this.command = command;
    }

    public void run() {
        Runtime.getRuntime().exec(command);
    }
}
```

An attacker crafts a serialized `CommandTask` (containing a
string like `cmd.exe`) and injects it into an input stream
that will be read by `CacheManager`. `CacheManager`
hydrates it into an object, invokes `run`, and the
attacker's managed to execute an arbitrary command.
Exploits in real life aren't likely to be this easy, but
they'll use the same mechanic.

## Defending yourself (#defense)

These types of bugs are _hopefully_ going to become less
common over time as language-agnostic serialization formats
like JSON that aren't as prone to this type of problem [2]
continue to gain in popularity, but that still leaves a lot
of existing software out there today.

The speakers above recommend the important vulnerabilities
to look for instances where an app is doing unsafe
deserialization. Trying to squash the problem by curbing
the available gadgets for exploitation is likely to be
fruitless because there's always more than will be found or
introduced, and apps tend to have sprawling class graphs
available in their runtime through libraries and transitive
dependencies.

A blunt defense might be to disable XML input entirely. The
technology's inherent complexity leads to more than it's
fair share of exploitations. This is quite reminiscent of
[CVE-2013-0156 in Rails][railsvuln] for example, which
allowed arbitrary code execution _through YAML embedded in
XML_ input.

[1] The full details of the breach are not yet known, and
I'm speculating based on [this communiqué][apacheresp] from
the Apache Struts Project Management Committee.

[2] I should add the caveat that you could build a flawed
JSON decoder that's prone to the same problem, but normally
these objects don't store anything like a class name, and
certainly won't by default.

[apacheresp]: https://blogs.apache.org/foundation/entry/apache-struts-statement-on-equifax
[metasploit]: https://github.com/rapid7/metasploit-framework/pull/8103
[railsvuln]: http://blog.codeclimate.com/blog/2013/01/10/rails-remote-code-execution-vulnerability-explained/
[struts]: https://struts.apache.org/
[strutsvuln]: https://cwiki.apache.org/confluence/display/WW/S2-052
[strutsvuln2]: https://nvd.nist.gov/vuln/detail/CVE-2017-5638
[talk]: https://frohoff.github.io/appseccali-marshalling-pickles/
