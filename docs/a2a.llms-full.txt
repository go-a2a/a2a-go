# https://a2a-protocol.org/latest/ llms-full.txt

<|firecrawl-page-1-lllmstxt|>
## Understanding A2A Protocol
[Skip to content](https://a2a-protocol.org/latest/topics/what-is-a2a/#what-is-a2a)

# What is A2A? [¶](https://a2a-protocol.org/latest/topics/what-is-a2a/\#what-is-a2a "Permanent link")

The Agent2Agent (A2A) Protocol is an open standard designed to solve a fundamental challenge in the rapidly evolving landscape of artificial intelligence: **how do AI agents, built by different teams, using different technologies, and owned by different organizations, communicate and collaborate effectively?**

As AI agents become more specialized and capable, the need for them to work together on complex tasks increases. Imagine a user asking their primary AI assistant to plan an international trip. This single request might involve coordinating the capabilities of several specialized agents:

1. An agent for flight bookings.
2. Another agent for hotel reservations.
3. A third for local tour recommendations and bookings.
4. A fourth to handle currency conversion and travel advisories.

Without a common communication protocol, integrating these diverse agents into a cohesive user experience is a significant engineering hurdle. Each integration would likely be a custom, point-to-point solution, making the system difficult to scale, maintain, and extend.

## The A2A Solution [¶](https://a2a-protocol.org/latest/topics/what-is-a2a/\#the-a2a-solution "Permanent link")

A2A provides a standardized way for these independent, often "opaque" (black-box) agentic systems to interact. It defines:

- **A common transport and format:** JSON-RPC 2.0 over HTTP(S) for how messages are structured and transmitted.
- **Discovery mechanisms (Agent Cards):** How agents can advertise their capabilities and be found by other agents.
- **Task management workflows:** How collaborative tasks are initiated, progressed, and completed. This includes support for tasks that may be long-running or require multiple turns of interaction.
- **Support for various data modalities:** How agents exchange not just text, but also files, structured data (like forms), and potentially other rich media.
- **Core principles for security and asynchronicity:** Guidelines for secure communication and handling tasks that might take significant time or involve human-in-the-loop processes.

## Key Design Principles of A2A [¶](https://a2a-protocol.org/latest/topics/what-is-a2a/\#key-design-principles-of-a2a "Permanent link")

The development of A2A is guided by several core principles:

- **Simplicity:** Leverage existing, well-understood standards like HTTP, JSON-RPC, and Server-Sent Events (SSE) where possible, rather than reinventing the wheel.
- **Enterprise Readiness:** Address critical enterprise needs such as authentication, authorization, security, privacy, tracing, and monitoring from the outset by aligning with standard web practices.
- **Asynchronous First:** Natively support long-running tasks and scenarios where agents or users might not be continuously connected, through mechanisms like streaming and push notifications.
- **Modality Agnostic:** Allow agents to communicate using a variety of content types, enabling rich and flexible interactions beyond plain text.
- **Opaque Execution:** Enable collaboration without requiring agents to expose their internal logic, memory, or proprietary tools. Agents interact based on declared capabilities and exchanged context, preserving intellectual property and enhancing security.

## Benefits of Using A2A [¶](https://a2a-protocol.org/latest/topics/what-is-a2a/\#benefits-of-using-a2a "Permanent link")

Adopting A2A can lead to significant advantages:

- **Increased Interoperability:** Break down silos between different AI agent ecosystems, allowing agents from various vendors and frameworks to work together.
- **Enhanced Agent Capabilities:** Allow developers to create more sophisticated applications by composing the strengths of multiple specialized agents.
- **Reduced Integration Complexity:** Standardize the "how" of agent communication, allowing teams to focus on the "what" – the value their agents provide.
- **Fostering Innovation:** Encourage the development of a richer ecosystem of specialized agents that can readily plug into larger collaborative workflows.
- **Future-Proofing:** Provide a flexible framework that can adapt as agent technologies continue to evolve.

By establishing common ground for agent-to-agent communication, A2A aims to accelerate the adoption and utility of AI agents across diverse industries and applications, paving the way for more powerful and collaborative AI systems.

## A2A Request Lifecycle [¶](https://a2a-protocol.org/latest/topics/what-is-a2a/\#a2a-request-lifecycle "Permanent link")

Next, learn about the [Key Concepts](https://a2a-protocol.org/latest/topics/key-concepts/) that form the foundation of the A2A protocol.

<|firecrawl-page-2-lllmstxt|>
## Agent Discovery in A2A
[Skip to content](https://a2a-protocol.org/latest/topics/agent-discovery/#agent-discovery-in-a2a)

# Agent Discovery in A2A [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#agent-discovery-in-a2a "Permanent link")

For AI agents to collaborate using the Agent2Agent (A2A) protocol, they first need to find each other and understand what capabilities the other agents offer. A2A standardizes the format of an agent's self-description through the **[Agent Card](https://a2a-protocol.org/latest/specification/#5-agent-discovery-the-agent-card)**. However, the methods for discovering these Agent Cards can vary depending on the environment and requirements.

## The Role of the Agent Card [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#the-role-of-the-agent-card "Permanent link")

The Agent Card is a JSON document that serves as a digital "business card" for an A2A Server (the remote agent). It is crucial for discovery and initiating interaction. Key information typically included in an Agent Card:

- **Identity:** `name`, `description`, `provider` information.
- **Service Endpoint:** The `url` where the A2A service can be reached.
- **A2A Capabilities:** Supported protocol features like `streaming` or `pushNotifications`.
- **Authentication:** Required authentication `schemes` (e.g., "Bearer", "OAuth2") to interact with the agent.
- **Skills:** A list of specific tasks or functions the agent can perform ( `AgentSkill` objects), including their `id`, `name`, `description`, `inputModes`, `outputModes`, and `examples`.

Client agents parse the Agent Card to determine if a remote agent is suitable for a given task, how to structure requests for its skills, and how to communicate with it securely.

## Discovery Strategies [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#discovery-strategies "Permanent link")

Here are common strategies for how a client agent might discover the Agent Card of a remote agent:

### 1\. Well-Known URI [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#1-well-known-uri "Permanent link")

This is a recommended approach for public agents or agents intended for broad discoverability within a specific domain.

- **Mechanism:** A2A Servers host their Agent Card at a standardized, "well-known" path on their domain.
- **Standard Path:** `https://{agent-server-domain}/.well-known/agent-card.json` (following the principles of [RFC 8615](https://www.ietf.org/rfc/rfc8615.txt) for well-known URIs).
- **Process:**
1. A client agent knows or programmatically discovers the domain of a potential A2A Server (e.g., `smart-thermostat.example.com`).
2. The client performs an HTTP `GET` request to `https://smart-thermostat.example.com/.well-known/agent-card.json`.
3. If the Agent Card exists and is accessible, the server returns it as a JSON response.
- **Advantages:** Simple, standardized, and enables automated discovery by crawlers or systems that can resolve domains. Effectively reduces the discovery problem to "find the agent's domain."
- **Considerations:** Best suited for agents intended for open discovery or discovery within an organization that controls the domain. The endpoint serving the Agent Card may itself require authentication if the card contains sensitive information.

### 2\. Curated Registries (Catalog-Based Discovery) [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#2-curated-registries-catalog-based-discovery "Permanent link")

For enterprise environments, marketplaces, or specialized ecosystems, Agent Cards can be published to and discovered via a central registry or catalog.

- **Mechanism:** An intermediary service (the registry) maintains a collection of Agent Cards. Clients query this registry to find agents based on various criteria (e.g., skills offered, tags, provider name, desired capabilities).
- **Process:**
1. A2A Servers (or their administrators) register their Agent Cards with the registry service. The mechanism for this registration is outside the scope of the A2A protocol itself.
2. Client agents query the registry's API (e.g., "find agents with 'image-generation' skill that support streaming").
3. The registry returns a list of matching Agent Cards or references to them.
- **Advantages:**
  - Centralized management, curation, and governance of available agents.
  - Facilitates discovery based on functional capabilities rather than just domain names.
  - Can implement access controls, policies, and trust mechanisms at the registry level.
  - Enables scenarios like company-specific or team-specific agent catalogs, or public marketplaces of A2A-compliant agents.
- **Considerations:** Requires an additional registry service. The A2A protocol does not currently define a standard API for such registries, though this is an area of potential future exploration and community standardization.

### 3\. Direct Configuration / Private Discovery [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#3-direct-configuration-private-discovery "Permanent link")

In many scenarios, especially within tightly coupled systems, for private agents, or during development and testing, clients might be directly configured with Agent Card information or a URL to fetch it.

- **Mechanism:** The client application has hardcoded Agent Card details, reads them from a local configuration file, receives them through an environment variable, or fetches them from a private, proprietary API endpoint known to the client.
- **Process:** This is highly specific to the application's deployment and configuration strategy.
- **Advantages:** Simple and effective for known, static relationships between agents or when dynamic discovery is not a requirement.
- **Considerations:** Less flexible for discovering new or updated agents dynamically. Changes to the remote agent's card might require re-configuration of the client. Proprietary API-based discovery is not standardized by A2A.

## Securing Agent Cards [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#securing-agent-cards "Permanent link")

Agent Cards themselves can sometimes contain information that should be protected, such as:

- The `url` of an internal-only or restricted-access agent.
- Details in the `authentication.credentials` field if it's used for scheme-specific, non-secret information (e.g., an OAuth token URL). Storing actual plaintext secrets in an Agent Card is **strongly discouraged**.
- Descriptions of sensitive or internal skills.

**Protection Mechanisms:**

- **Access Control on the Endpoint:** The HTTP endpoint serving the Agent Card (whether it's the `/.well-known/agent-card.json` path, a registry API, or a custom URL) should be secured using standard web practices if the card is not intended for public, unauthenticated access.
  - **mTLS:** Require mutual TLS for client authentication if appropriate for the trust model.
  - **Network Restrictions:** Limit access to specific IP ranges, VPCs, or private networks.
  - **Authentication:** Require standard HTTP authentication (e.g., OAuth 2.0 Bearer token, API Key) to access the Agent Card itself.
- **Selective Disclosure by Registries:** Agent registries can implement logic to return different Agent Cards or varying levels of detail based on the authenticated client's identity and permissions. For example, a public query might return a limited card, while an authenticated partner query might receive a card with more details.

It's crucial to remember that if an Agent Card were to contain sensitive data (again, **not recommended** for secrets), the card itself **must never** be available without strong authentication and authorization. The A2A protocol encourages authentication schemes where the client obtains dynamic credentials out-of-band, rather than relying on static secrets embedded in the Agent Card.

## Future Considerations [¶](https://a2a-protocol.org/latest/topics/agent-discovery/\#future-considerations "Permanent link")

The A2A community may explore standardizing aspects of registry interactions or more advanced, semantic discovery protocols in the future. Feedback and contributions in this area are welcome to enhance the discoverability and interoperability of A2A agents.

<|firecrawl-page-3-lllmstxt|>
## Next Steps in A2A
[Skip to content](https://a2a-protocol.org/latest/tutorials/python/8-next-steps/#next-steps)

# Next Steps [¶](https://a2a-protocol.org/latest/tutorials/python/8-next-steps/\#next-steps "Permanent link")

Congratulations on completing the A2A Python SDK Tutorial! You've learned how to:

- Set up your environment for A2A development.
- Define Agent Skills and Agent Cards using the SDK's types.
- Implement a basic HelloWorld A2A server and client.
- Understand and implement streaming capabilities.
- Integrate a more complex agent using LangGraph, demonstrating task state management and tool use.

You now have a solid foundation for building and integrating your own A2A-compliant agents.

## Where to Go From Here? [¶](https://a2a-protocol.org/latest/tutorials/python/8-next-steps/\#where-to-go-from-here "Permanent link")

Here are some ideas and resources to continue your A2A journey:

- **Explore Other Examples:**
  - Check out the other examples in the `a2a-samples/samples/` directory in the [A2A GitHub repository](https://github.com/a2aproject/a2a-samples/tree/main/samples) for more complex agent integrations and features.
  - The main A2A repository also has [samples for other languages and frameworks](https://github.com/a2aproject/A2A/tree/main/samples).
- **Deepen Your Protocol Understanding:**
  - 📚 Read the complete [A2A Protocol Documentation site](https://google.github.io/A2A/) for a comprehensive overview.
  - 📝 Review the detailed [A2A Protocol Specification](https://a2a-protocol.org/latest/specification/) to understand the nuances of all data structures and RPC methods.
- **Review Key A2A Topics:**
  - [A2A and MCP](https://a2a-protocol.org/latest/topics/a2a-and-mcp/): Understand how A2A complements the Model Context Protocol for tool usage.
  - [Enterprise-Ready Features](https://a2a-protocol.org/latest/topics/enterprise-ready/): Learn about security, observability, and other enterprise considerations.
  - [Streaming & Asynchronous Operations](https://a2a-protocol.org/latest/topics/streaming-and-async/): Get more details on SSE and push notifications.
  - [Agent Discovery](https://a2a-protocol.org/latest/topics/agent-discovery/): Explore different ways agents can find each other.
- **Build Your Own Agent:**
  - Try creating a new A2A agent using your favorite Python agent framework (like LangChain, CrewAI, AutoGen, Semantic Kernel, or a custom solution).
  - Implement the `a2a.server.AgentExecutor` interface to bridge your agent's logic with the A2A protocol.
  - Think about what unique skills your agent could offer and how its Agent Card would represent them.
- **Experiment with Advanced Features:**
  - Implement robust task management with a persistent `TaskStore` if your agent handles long-running or multi-session tasks.
  - Explore implementing push notifications if your agent's tasks are very long-lived.
  - Consider more complex input and output modalities (e.g., handling file uploads/downloads, or structured data via `DataPart`).
- **Contribute to the A2A Community:**
  - Join the discussions on the [A2A GitHub Discussions page](https://github.com/a2aproject/A2A/discussions).
  - Report issues or suggest improvements via [GitHub Issues](https://github.com/a2aproject/A2A/issues).
  - Consider contributing code, examples, or documentation. See the [CONTRIBUTING.md](https://github.com/a2aproject/A2A/blob/main/CONTRIBUTING.md) guide.

The A2A protocol aims to foster an ecosystem of interoperable AI agents. By building and sharing A2A-compliant agents, you can be a part of this exciting development!

<|firecrawl-page-4-lllmstxt|>
## Agent Skills and Card
[Skip to content](https://a2a-protocol.org/latest/tutorials/python/3-agent-skills-and-card/#3-agent-skills-agent-card)

# 3\. Agent Skills & Agent Card [¶](https://a2a-protocol.org/latest/tutorials/python/3-agent-skills-and-card/\#3-agent-skills-agent-card "Permanent link")

Before an A2A agent can do anything, it needs to define what it _can_ do (its skills) and how other agents or clients can find out about these capabilities (its Agent Card).

We'll use the `helloworld` example located in [`a2a-samples/samples/python/agents/helloworld/`](https://github.com/a2aproject/a2a-samples/tree/main/samples/python/agents/helloworld).

## Agent Skills [¶](https://a2a-protocol.org/latest/tutorials/python/3-agent-skills-and-card/\#agent-skills "Permanent link")

An **Agent Skill** describes a specific capability or function the agent can perform. It's a building block that tells clients what kinds of tasks the agent is good for.

Key attributes of an `AgentSkill` (defined in `a2a.types`):

- `id`: A unique identifier for the skill.
- `name`: A human-readable name.
- `description`: A more detailed explanation of what the skill does.
- `tags`: Keywords for categorization and discovery.
- `examples`: Sample prompts or use cases.
- `inputModes` / `outputModes`: Supported Media Types for input and output (e.g., "text/plain", "application/json").

In `__main__.py`, you can see how a skill for the Helloworld agent is defined:

```md-code__content
skill = AgentSkill(
    id='hello_world',
    name='Returns hello world',
    description='just returns hello world',
    tags=['hello world'],
    examples=['hi', 'hello world'],
)

```

This skill is very simple: it's named "Returns hello world" and primarily deals with text.

## Agent Card [¶](https://a2a-protocol.org/latest/tutorials/python/3-agent-skills-and-card/\#agent-card "Permanent link")

The **Agent Card** is a JSON document that an A2A Server makes available, typically at a `.well-known/agent-card.json` endpoint. It's like a digital business card for the agent.

Key attributes of an `AgentCard` (defined in `a2a.types`):

- `name`, `description`, `version`: Basic identity information.
- `url`: The endpoint where the A2A service can be reached.
- `capabilities`: Specifies supported A2A features like `streaming` or `pushNotifications`.
- `defaultInputModes` / `defaultOutputModes`: Default Media Types for the agent.
- `skills`: A list of `AgentSkill` objects that the agent offers.

The `helloworld` example defines its Agent Card like this:

```md-code__content
# This will be the public-facing agent card
public_agent_card = AgentCard(
    name='Hello World Agent',
    description='Just a hello world agent',
    url='http://localhost:9999/',
    version='1.0.0',
    default_input_modes=['text'],
    default_output_modes=['text'],
    capabilities=AgentCapabilities(streaming=True),
    skills=[skill],  # Only the basic skill for the public card
    supports_authenticated_extended_card=True,
)

```

This card tells us the agent is named "Hello World Agent", runs at `http://localhost:9999/`, supports text interactions, and has the `hello_world` skill. It also indicates public authentication, meaning no specific credentials are required.

Understanding the Agent Card is crucial because it's how a client discovers an agent and learns how to interact with it.

<|firecrawl-page-5-lllmstxt|>
## A2A Protocol Roadmap
[Skip to content](https://a2a-protocol.org/latest/roadmap/#a2a-protocol-roadmap)

# A2A protocol roadmap [¶](https://a2a-protocol.org/latest/roadmap/\#a2a-protocol-roadmap "Permanent link")

**Last updated:** Jul 16, 2025

## Near-term initiatives [¶](https://a2a-protocol.org/latest/roadmap/\#near-term-initiatives "Permanent link")

- Release `0.3` version of the protocol which we intend to keep supported and without breaking changes for a significant amount of time with backward compatibility of the SDKs starting at version `0.3`. As part of this release there are a few known breaking changes including:
  - Update the `/.well-known/agent.json` path for hosting Agent Cards to `/.well-known/agent-card.json` based on feedback from IANA.
  - Refactor class fields to be more Pythonic and adopt `snake_case`. [PR 199](https://github.com/a2aproject/a2a-python/pull/199)
- Solidify the support for [A2A extensions](https://a2a-protocol.org/latest/topics/extensions/) with SDK support (starting with the Python SDK) and publishing sample extensions.
- Introduce support for signed Agent Cards [Discussion 199](https://github.com/a2aproject/A2A/discussions/199#discussioncomment-13770576) to allow verifying the integrity of Agent Card content.
- Enhance the client side support in SDK (starting with Python) to expose ready-to-use A2A clients, streamlined auth handling and improved handling of tasks.

To review recent protocol changes see [Release Notes](https://github.com/a2aproject/A2A/releases).

## Longer term (3-6 month period) roadmap [¶](https://a2a-protocol.org/latest/roadmap/\#longer-term-3-6-month-period-roadmap "Permanent link")

### Governance [¶](https://a2a-protocol.org/latest/roadmap/\#governance "Permanent link")

The protocol has been [donated](https://www.linuxfoundation.org/press/linux-foundation-launches-the-agent2agent-protocol-project-to-enable-secure-intelligent-communication-between-ai-agents) to the Linux Foundation. The TSC is working on implementing a governance structure that prioritizes community-led development with standardized processes for contributing to the specification, SDKs and tooling. As part of the effort there will be dedicated working groups created for specific areas of the protocol.

### Agent Registry [¶](https://a2a-protocol.org/latest/roadmap/\#agent-registry "Permanent link")

Agent Registry enables the discovery of agents and is a critical component of a multi-agent system. There is an active and ongoing discussion in the community around the latest [Discussion 741](https://github.com/a2aproject/A2A/discussions/741).

### Validation [¶](https://a2a-protocol.org/latest/roadmap/\#validation "Permanent link")

As the A2A ecosystem matures, it becomes critical for the A2A community to have tools to validate their agents. The community has launched two efforts to help with validation which the group will continue to enhance in the coming months. Learn more about [A2A Inspector](https://github.com/a2aproject/a2a-inspector) and the [A2A Protocol Technology Compatibility Kit](https://github.com/a2aproject/a2a-tck) (TCK).

### SDKs [¶](https://a2a-protocol.org/latest/roadmap/\#sdks "Permanent link")

A2A Project currently hosts SDKs in four languages (Python, JS, Java, .NET) and contributors are adding more including Go (in progress).

### Community best practices [¶](https://a2a-protocol.org/latest/roadmap/\#community-best-practices "Permanent link")

As companies and individuals deploy A2A systems at an increasing pace, we are looking to accelerate the learning of the community by collecting and sharing the best practices and success stories that A2A enabled.

<|firecrawl-page-6-lllmstxt|>
## A2A Protocol Sitemap
https://a2a-protocol.org/dev/2025-07-29https://a2a-protocol.org/dev/community/2025-07-29https://a2a-protocol.org/dev/partners/2025-07-29https://a2a-protocol.org/dev/roadmap/2025-07-29https://a2a-protocol.org/dev/specification/2025-07-29https://a2a-protocol.org/dev/sdk/python/2025-07-29https://a2a-protocol.org/dev/topics/a2a-and-mcp/2025-07-29https://a2a-protocol.org/dev/topics/agent-discovery/2025-07-29https://a2a-protocol.org/dev/topics/enterprise-ready/2025-07-29https://a2a-protocol.org/dev/topics/extensions/2025-07-29https://a2a-protocol.org/dev/topics/key-concepts/2025-07-29https://a2a-protocol.org/dev/topics/life-of-a-task/2025-07-29https://a2a-protocol.org/dev/topics/streaming-and-async/2025-07-29https://a2a-protocol.org/dev/topics/what-is-a2a/2025-07-29https://a2a-protocol.org/dev/tutorials/python/1-introduction/2025-07-29https://a2a-protocol.org/dev/tutorials/python/2-setup/2025-07-29https://a2a-protocol.org/dev/tutorials/python/3-agent-skills-and-card/2025-07-29https://a2a-protocol.org/dev/tutorials/python/4-agent-executor/2025-07-29https://a2a-protocol.org/dev/tutorials/python/5-start-server/2025-07-29https://a2a-protocol.org/dev/tutorials/python/6-interact-with-server/2025-07-29https://a2a-protocol.org/dev/tutorials/python/7-streaming-and-multiturn/2025-07-29https://a2a-protocol.org/dev/tutorials/python/8-next-steps/2025-07-29

<urlsetxmlns="http://www.sitemaps.org/schemas/sitemap/0.9">

<url>

<loc>https://a2a-protocol.org/dev/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/community/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/partners/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/roadmap/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/specification/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/sdk/python/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/a2a-and-mcp/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/agent-discovery/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/enterprise-ready/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/extensions/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/key-concepts/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/life-of-a-task/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/streaming-and-async/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/topics/what-is-a2a/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/1-introduction/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/2-setup/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/3-agent-skills-and-card/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/4-agent-executor/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/5-start-server/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/6-interact-with-server/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/7-streaming-and-multiturn/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

<url>

<loc>https://a2a-protocol.org/dev/tutorials/python/8-next-steps/</loc>

<lastmod>2025-07-29</lastmod>

...

</url>

...

</urlset>

<|firecrawl-page-7-lllmstxt|>
## Interacting with A2A Server
[Skip to content](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/#6-interacting-with-the-server)

# 6\. Interacting with the Server [¶](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/\#6-interacting-with-the-server "Permanent link")

With the Helloworld A2A server running, let's send some requests to it. The SDK includes a client ( `A2AClient`) that simplifies these interactions.

## The Helloworld Test Client [¶](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/\#the-helloworld-test-client "Permanent link")

The `test_client.py` script demonstrates how to:

1. Fetch the Agent Card from the server.
2. Create an `A2AClient` instance.
3. Send both non-streaming ( `message/send`) and streaming ( `message/stream`) requests.

Open a **new terminal window**, activate your virtual environment, and navigate to the `a2a-samples` directory.

Activate virtual environment (Be sure to do this in the same directory where you created the virtual environment):

[Mac/Linux](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/#maclinux)[Windows](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/#windows)

```md-code__content
source .venv/bin/activate

```

```md-code__content
.venv\Scripts\activate

```

Run the test client:

```md-code__content
# from the a2a-samples directory
python samples/python/agents/helloworld/test_client.py

```

## Understanding the Client Code [¶](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/\#understanding-the-client-code "Permanent link")

Let's look at key parts of `test_client.py`:

1. **Fetching the Agent Card & Initializing the Client**:



```md-code__content
base_url = 'http://localhost:9999'

async with httpx.AsyncClient() as httpx_client:
       # Initialize A2ACardResolver
       resolver = A2ACardResolver(
           httpx_client=httpx_client,
           base_url=base_url,
           # agent_card_path uses default, extended_agent_card_path also uses default
       )

```



The `A2ACardResolver` class is a convenience. It first fetches the `AgentCard` from the server's `/.well-known/agent-card.json` endpoint (based on the provided base URL) and then initializes the client with it.

2. **Sending a Non-Streaming Message ( `send_message`)**:



```md-code__content
client = A2AClient(
       httpx_client=httpx_client, agent_card=final_agent_card_to_use
)
logger.info('A2AClient initialized.')

send_message_payload: dict[str, Any] = {
       'message': {
           'role': 'user',
           'parts': [\
               {'kind': 'text', 'text': 'how much is 10 USD in INR?'}\
           ],
           'messageId': uuid4().hex,
       },
}
request = SendMessageRequest(
       id=str(uuid4()), params=MessageSendParams(**send_message_payload)
)

response = await client.send_message(request)
print(response.model_dump(mode='json', exclude_none=True))

```


   - The `send_message_payload` constructs the data for `MessageSendParams`.
   - This is wrapped in a `SendMessageRequest`.
   - It includes a `message` object with the `role` set to "user" and the content in `parts`.
   - The Helloworld agent's `execute` method will enqueue a single "Hello World" message. The `DefaultRequestHandler` will retrieve this and send it as the response.
   - The `response` will be a `SendMessageResponse` object, which contains either a `SendMessageSuccessResponse` (with the agent's `Message` as the result) or a `JSONRPCErrorResponse`.
3. **Handling Task IDs (Illustrative Note for Helloworld)**:

The Helloworld client ( `test_client.py`) doesn't attempt `get_task` or `cancel_task` directly because the simple Helloworld agent's `execute` method, when called via `message/send`, results in the `DefaultRequestHandler` returning a direct `Message` response rather than a `Task` object. More complex agents that explicitly manage tasks (like the LangGraph example) would return a `Task` object from `message/send`, and its `id` could then be used for `get_task` or `cancel_task`.

4. **Sending a Streaming Message ( `send_message_streaming`)**:



```md-code__content
streaming_request = SendStreamingMessageRequest(
       id=str(uuid4()), params=MessageSendParams(**send_message_payload)
)

stream_response = client.send_message_streaming(streaming_request)

async for chunk in stream_response:
       print(chunk.model_dump(mode='json', exclude_none=True))

```


   - This method calls the agent's `message/stream` endpoint. The `DefaultRequestHandler` will invoke the `HelloWorldAgentExecutor.execute` method.
   - The `execute` method enqueues one "Hello World" message, and then the event queue is closed.
   - The client will receive this single message as one `SendStreamingMessageResponse` event, and then the stream will terminate.
   - The `stream_response` is an `AsyncGenerator`.

## Expected Output [¶](https://a2a-protocol.org/latest/tutorials/python/6-interact-with-server/\#expected-output "Permanent link")

When you run `test_client.py`, you'll see JSON outputs for:

- The non-streaming response (a single "Hello World" message).
- The streaming response (a single "Hello World" message as one chunk, after which the stream ends).

The `id` fields in the output will vary with each run.

```md-code__content
// Non-streaming response
{"jsonrpc":"2.0","id":"xxxxxxxx","result":{"type":"message","role":"agent","parts":[{"type":"text","text":"Hello World"}],"messageId":"yyyyyyyy"}}
// Streaming response (one chunk)
{"jsonrpc":"2.0","id":"zzzzzzzz","result":{"type":"message","role":"agent","parts":[{"type":"text","text":"Hello World"}],"messageId":"wwwwwwww","final":true}}

```

_(Actual IDs like `xxxxxxxx`, `yyyyyyyy`, `zzzzzzzz`, `wwwwwwww` will be different UUIDs/request IDs)_

This confirms your server is correctly handling basic A2A interactions with the updated SDK structure!

Now you can shut down the server by typing Ctrl+C in the terminal window where `__main__.py` is running.

<|firecrawl-page-8-lllmstxt|>
## Agent Executor Tutorial
[Skip to content](https://a2a-protocol.org/latest/tutorials/python/4-agent-executor/#4-the-agent-executor)

# 4\. The Agent Executor [¶](https://a2a-protocol.org/latest/tutorials/python/4-agent-executor/\#4-the-agent-executor "Permanent link")

The core logic of how an A2A agent processes requests and generates responses/events is handled by an **Agent Executor**. The A2A Python SDK provides an abstract base class `a2a.server.agent_execution.AgentExecutor` that you implement.

## `AgentExecutor` Interface [¶](https://a2a-protocol.org/latest/tutorials/python/4-agent-executor/\#agentexecutor-interface "Permanent link")

The `AgentExecutor` class defines two primary methods:

- `async def execute(self, context: RequestContext, event_queue: EventQueue)`: Handles incoming requests that expect a response or a stream of events. It processes the user's input (available via `context`) and uses the `event_queue` to send back `Message`, `Task`, `TaskStatusUpdateEvent`, or `TaskArtifactUpdateEvent` objects.
- `async def cancel(self, context: RequestContext, event_queue: EventQueue)`: Handles requests to cancel an ongoing task.

The `RequestContext` provides information about the incoming request, such as the user's message and any existing task details. The `EventQueue` is used by the executor to send events back to the client.

## Helloworld Agent Executor [¶](https://a2a-protocol.org/latest/tutorials/python/4-agent-executor/\#helloworld-agent-executor "Permanent link")

Let's look at `agent_executor.py`. It defines `HelloWorldAgentExecutor`.

1. **The Agent ( `HelloWorldAgent`)**:
    This is a simple helper class that encapsulates the actual "business logic".



```md-code__content
class HelloWorldAgent:
       """Hello World Agent."""

       async def invoke(self) -> str:
           return 'Hello World'

```



It has a simple `invoke` method that returns the string "Hello World".

2. **The Executor ( `HelloWorldAgentExecutor`)**:
    This class implements the `AgentExecutor` interface.
   - **`__init__`**:



     ```md-code__content
     class HelloWorldAgentExecutor(AgentExecutor):
         """Test AgentProxy Implementation."""

         def __init__(self):
             self.agent = HelloWorldAgent()

     ```



     It instantiates the `HelloWorldAgent`.

   - **`execute`**:



     ```md-code__content
     async def execute(
         self,
         context: RequestContext,
         event_queue: EventQueue,
     ) -> None:
         result = await self.agent.invoke()
         await event_queue.enqueue_event(new_agent_text_message(result))

     ```



     When a `message/send` or `message/stream` request comes in (both are handled by `execute` in this simplified executor):
     1. It calls `self.agent.invoke()` to get the "Hello World" string.
     2. It creates an A2A `Message` object using the `new_agent_text_message` utility function.
     3. It enqueues this message onto the `event_queue`. The underlying `DefaultRequestHandler` will then process this queue to send the response(s) to the client. For a single message like this, it will result in a single response for `message/send` or a single event for `message/stream` before the stream closes.
   - **`cancel`**:
      The Helloworld example's `cancel` method simply raises an exception, indicating that cancellation is not supported for this basic agent.



     ```md-code__content
     async def cancel(
         self, context: RequestContext, event_queue: EventQueue
     ) -> None:
         raise Exception('cancel not supported')

     ```

The `AgentExecutor` acts as the bridge between the A2A protocol (managed by the request handler and server application) and your agent's specific logic. It receives context about the request and uses an event queue to communicate results or updates back.

<|firecrawl-page-9-lllmstxt|>
## A2A and MCP Protocols
[Skip to content](https://a2a-protocol.org/latest/topics/a2a-and-mcp/#a2a-and-mcp-complementary-protocols-for-agentic-systems)

# A2A and MCP: Complementary Protocols for Agentic Systems [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#a2a-and-mcp-complementary-protocols-for-agentic-systems "Permanent link")

## A2A ❤️ MCP [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#a2a-mcp "Permanent link")

In the landscape of AI agent development, two key types of protocols are emerging to facilitate interoperability: those for connecting agents to **tools and resources**, and those for enabling **agent-to-agent collaboration**. The Agent2Agent (A2A) Protocol and the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) address these distinct but related needs.

**TL;DR;** Agentic applications need both A2A and MCP. We recommend MCP for tools and A2A for agents.

## Why Different Protocols? [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#why-different-protocols "Permanent link")

The distinction arises from the nature of what an agent interacts with:

- **Tools & Resources:**
  - These are typically primitives with well-defined, structured inputs and outputs. They perform specific, often stateless, functions (e.g., a calculator, a database query API, a weather lookup service).
  - Their behavior is generally predictable and transactional.
  - Interaction is often a single request-response cycle.
- **Agents:**
  - These are more autonomous systems. They can reason, plan, use multiple tools, maintain state over longer interactions, and engage in complex, often multi-turn dialogues to achieve novel or evolving tasks.
  - Their behavior can be emergent and less predictable than a simple tool.
  - Interaction often involves ongoing tasks, context sharing, and negotiation.

Agentic applications need to leverage both: agents use tools to gather information and perform actions, and agents collaborate with other agents to tackle broader, more complex goals.

## Model Context Protocol (MCP) [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#model-context-protocol-mcp "Permanent link")

- **Focus:** MCP standardizes how AI models and agents connect to and interact with **tools, APIs, data sources, and other external resources.**
- **Mechanism:** It defines a structured way to describe tool capabilities (akin to function calling in Large Language Models), pass inputs to them, and receive structured outputs.
- **Use Cases:**
  - Enabling an LLM to call an external API (e.g., fetch current stock prices).
  - Allowing an agent to query a database with specific parameters.
  - Connecting an agent to a set of predefined functions or services.
- **Ecosystem:** MCP aims to create an ecosystem where tool providers can easily expose their services to various AI models and agent frameworks, and agent developers can easily consume these tools in a standardized way.

## Agent2Agent Protocol (A2A) [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#agent2agent-protocol-a2a "Permanent link")

- **Focus:** A2A standardizes how independent, often opaque, **AI agents communicate and collaborate with each other as peers.**
- **Mechanism:** It provides an application-level protocol for agents to:
  - Discover each other's high-level skills and capabilities (via Agent Cards).
  - Negotiate interaction modalities (text, files, structured data).
  - Manage shared, stateful, and potentially long-running tasks.
  - Exchange conversational context, instructions, and complex, multi-part results.
- **Use Cases:**
  - A customer service agent delegating a complex billing inquiry to a specialized billing agent, maintaining context of the customer interaction.
  - A travel planning agent coordinating with separate flight, hotel, and activity booking agents, managing a multi-stage booking process.
  - Agents exchanging information and status updates for a collaborative project that evolves over time.
- **Key Difference from Tool Interaction:** A2A allows for more dynamic, stateful, and potentially multi-modal interactions than typically seen with simple tool calls. Agents using A2A communicate _as agents_ (or on behalf of users) rather than just invoking a discrete function.

## How A2A and MCP Complement Each Other [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#how-a2a-and-mcp-complement-each-other "Permanent link")

A2A and MCP are not mutually exclusive; they are highly complementary and address different layers of an agentic system's interaction needs.

![Diagram showing A2A and MCP working together. A User interacts with Agent A via A2A. Agent A interacts with Agent B via A2A. Agent B uses MCP to interact with Tool 1 and Tool 2.](https://a2a-protocol.org/latest/assets/a2a-mcp.png)

_An agentic application might use A2A to communicate with other agents, while each agent internally uses MCP to interact with its specific tools and resources._

### Example Scenario: The Auto Repair Shop [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#example-scenario-the-auto-repair-shop "Permanent link")

> Consider an auto repair shop staffed by autonomous AI agent "mechanics" who use special-purpose tools (such as vehicle jacks, multimeters, and socket wrenches) to diagnose and repair problems. The workers often have to diagnose and repair problems they have not seen before. The repair process can involve extensive conversations with a customer, research, and working with part suppliers.

1. **Customer Interaction (User-to-Agent via A2A):**
   - A customer (or their primary assistant agent) uses A2A to communicate with the "Shop Manager" agent: _"My car is making a rattling noise."_
   - The Shop Manager agent uses A2A for a multi-turn diagnostic conversation: _"Can you send a video of the noise?"_, _"I see some fluid leaking. How long has this been happening?"_
2. **Internal Tool Usage (Agent-to-Tool via MCP):**
   - The Mechanic agent, assigned the task by the Shop Manager, needs to diagnose the issue. It uses MCP to interact with its specialized tools:
     - MCP call to a "Vehicle Diagnostic Scanner" tool: `scan_vehicle_for_error_codes(vehicle_id='XYZ123')`.
     - MCP call to a "Repair Manual Database" tool: `get_repair_procedure(error_code='P0300', vehicle_make='Toyota', vehicle_model='Camry')`.
     - MCP call to a "Platform Lift" tool: `raise_platform(height_meters=2)`.
3. **Supplier Interaction (Agent-to-Agent via A2A):**
   - The Mechanic agent determines a specific part is needed. It uses A2A to communicate with a "Parts Supplier" agent: _"Do you have part #12345 in stock for a Toyota Camry 2018?"_
   - The Parts Supplier agent, also an A2A-compliant system, responds, potentially leading to an order.

In this example:

- **A2A** facilitates the higher-level, conversational, and task-oriented interactions between the customer and the shop, and between the shop's agents and external supplier agents.
- **MCP** enables the mechanic agent to use its specific, structured tools to perform its diagnostic and repair functions.

## Representing A2A Agents as MCP Resources [¶](https://a2a-protocol.org/latest/topics/a2a-and-mcp/\#representing-a2a-agents-as-mcp-resources "Permanent link")

It's conceivable that an A2A Server (a remote agent) could also expose some of its skills as MCP-compatible resources, especially if those skills are well-defined and can be invoked in a more tool-like, stateless manner. In such a case, another agent might "discover" this A2A agent's specific skill via an MCP-style tool description (perhaps derived from its Agent Card).

However, the primary strength of A2A lies in its support for more flexible, stateful, and collaborative interactions that go beyond typical tool invocation. A2A is about agents _partnering_ on tasks, while MCP is more about agents _using_ capabilities.

By leveraging both A2A for inter-agent collaboration and MCP for tool integration, developers can build more powerful, flexible, and interoperable AI systems.

<|firecrawl-page-10-lllmstxt|>
## A2A Protocol Specification
[Skip to content](https://a2a-protocol.org/latest/specification/#agent2agent-a2a-protocol-specification)

# Agent2Agent (A2A) Protocol Specification [¶](https://a2a-protocol.org/latest/specification/\#agent2agent-a2a-protocol-specification "Permanent link")

**Version:** `dev`

See [Release Notes](https://github.com/a2aproject/A2A/releases) for changes made between versions.

## 1\. Introduction [¶](https://a2a-protocol.org/latest/specification/\#1-introduction "Permanent link")

The Agent2Agent (A2A) Protocol is an open standard designed to facilitate communication and interoperability between independent, potentially opaque AI agent systems. In an ecosystem where agents might be built using different frameworks, languages, or by different vendors, A2A provides a common language and interaction model.

This document provides the detailed technical specification for the A2A protocol. Its primary goal is to enable agents to:

- Discover each other's capabilities.
- Negotiate interaction modalities (text, files, structured data).
- Manage collaborative tasks.
- Securely exchange information to achieve user goals **without needing access to each other's internal state, memory, or tools.**

### 1.1. Key Goals of A2A [¶](https://a2a-protocol.org/latest/specification/\#11-key-goals-of-a2a "Permanent link")

- **Interoperability:** Bridge the communication gap between disparate agentic systems.
- **Collaboration:** Enable agents to delegate tasks, exchange context, and work together on complex user requests.
- **Discovery:** Allow agents to dynamically find and understand the capabilities of other agents.
- **Flexibility:** Support various interaction modes including synchronous request/response, streaming for real-time updates, and asynchronous push notifications for long-running tasks.
- **Security:** Facilitate secure communication patterns suitable for enterprise environments, relying on standard web security practices.
- **Asynchronicity:** Natively support long-running tasks and interactions that may involve human-in-the-loop scenarios.

### 1.2. Guiding Principles [¶](https://a2a-protocol.org/latest/specification/\#12-guiding-principles "Permanent link")

- **Simple:** Reuse existing, well-understood standards (HTTP, JSON-RPC 2.0, Server-Sent Events).
- **Enterprise Ready:** Address authentication, authorization, security, privacy, tracing, and monitoring by aligning with established enterprise practices.
- **Async First:** Designed for (potentially very) long-running tasks and human-in-the-loop interactions.
- **Modality Agnostic:** Support exchange of diverse content types including text, audio/video (via file references), structured data/forms, and potentially embedded UI components (e.g., iframes referenced in parts).
- **Opaque Execution:** Agents collaborate based on declared capabilities and exchanged information, without needing to share their internal thoughts, plans, or tool implementations.

For a broader understanding of A2A's purpose and benefits, see [What is A2A?](https://a2a-protocol.org/latest/topics/what-is-a2a/).

## 2\. Core Concepts Summary [¶](https://a2a-protocol.org/latest/specification/\#2-core-concepts-summary "Permanent link")

A2A revolves around several key concepts. For detailed explanations, please refer to the [Key Concepts guide](https://a2a-protocol.org/latest/topics/key-concepts/).

- **A2A Client:** An application or agent that initiates requests to an A2A Server on behalf of a user or another system.
- **A2A Server (Remote Agent):** An agent or agentic system that exposes an A2A-compliant HTTP endpoint, processing tasks and providing responses.
- **Agent Card:** A JSON metadata document published by an A2A Server, describing its identity, capabilities, skills, service endpoint, and authentication requirements.
- **Message:** A communication turn between a client and a remote agent, having a `role` ("user" or "agent") and containing one or more `Parts`.
- **Task:** The fundamental unit of work managed by A2A, identified by a unique ID. Tasks are stateful and progress through a defined lifecycle.
- **Part:** The smallest unit of content within a Message or Artifact (e.g., `TextPart`, `FilePart`, `DataPart`).
- **Artifact:** An output (e.g., a document, image, structured data) generated by the agent as a result of a task, composed of `Parts`.
- **Streaming (SSE):** Real-time, incremental updates for tasks (status changes, artifact chunks) delivered via Server-Sent Events.
- **Push Notifications:** Asynchronous task updates delivered via server-initiated HTTP POST requests to a client-provided webhook URL, for long-running or disconnected scenarios.
- **Context:** An optional, server-generated identifier to logically group related tasks.
- **Extension:** A mechanism for agents to provide additional functionality or data beyond the core A2A specification.

## 3\. Transport and Format [¶](https://a2a-protocol.org/latest/specification/\#3-transport-and-format "Permanent link")

### 3.1. Transport Layer Requirements [¶](https://a2a-protocol.org/latest/specification/\#31-transport-layer-requirements "Permanent link")

A2A supports multiple transport protocols, all operating over **HTTP(S)**. Agents have flexibility in choosing which transport protocols to implement based on their specific requirements and use cases:

- A2A communication **MUST** occur over **HTTP(S)**.
- The A2A Server exposes its service at one or more URLs defined in its `AgentCard`.
- Agents **MUST** implement at least one of the three core transport protocols defined in this specification.
- All supported transport protocols are considered equal in status and capability.

### 3.2. Supported Transport Protocols [¶](https://a2a-protocol.org/latest/specification/\#32-supported-transport-protocols "Permanent link")

A2A defines three core transport protocols. **A2A-compliant agents SHOULD implement at least one of these transport protocols. They MAY be compliant implementing a transport extension as defined in [3.2.4](https://a2a-protocol.org/latest/specification/#324-transport-extensions)** All three protocols are considered equal in status, and agents may choose to implement any combination of them based on their requirements.

#### 3.2.1. JSON-RPC 2.0 Transport [¶](https://a2a-protocol.org/latest/specification/\#321-json-rpc-20-transport "Permanent link")

Agents **MAY** support JSON-RPC 2.0 transport. If implemented, it **MUST** conform to these requirements:

- The primary data format is **[JSON-RPC 2.0](https://www.jsonrpc.org/specification)** for all requests and responses (excluding SSE stream wrapper).
- Client requests and server responses **MUST** adhere to the JSON-RPC 2.0 specification.
- The `Content-Type` header for HTTP requests and responses containing JSON-RPC payloads **MUST** be `application/json`.
- Method names follow the pattern `{category}/{action}` (e.g., `"message/send"`, `"tasks/get"`).

#### 3.2.2. gRPC Transport [¶](https://a2a-protocol.org/latest/specification/\#322-grpc-transport "Permanent link")

Agents **MAY** support gRPC transport. If implemented, it **MUST** conform to these requirements:

- **Protocol Definition**: **MUST** use the normative Protocol Buffers definition in [`specification/grpc/a2a.proto`](https://a2a-protocol.org/latest/specification/specification/grpc/a2a.proto).
- **Message Serialization**: **MUST** use Protocol Buffers version 3 for message serialization.
- **Service Definition**: **MUST** implement the `A2AService` gRPC service as defined in the proto file.
- **Method Coverage**: **MUST** provide all methods with functionally equivalent behavior to other supported transports.
- **Field Mapping**: **MUST** use the `json_name` annotations for HTTP/JSON transcoding compatibility.
- **Error Handling**: **MUST** map A2A error codes to appropriate gRPC status codes as defined in the proto annotations.
- **Transport Security**: **MUST** support TLS encryption (gRPC over HTTP/2 with TLS).

#### 3.2.3. HTTP+JSON/REST Transport [¶](https://a2a-protocol.org/latest/specification/\#323-httpjsonrest-transport "Permanent link")

Agents **MAY** support REST-style HTTP+JSON transport. If implemented, it **MUST** conform to these requirements:

- **HTTP Methods**: **MUST** use appropriate HTTP verbs (GET for queries, POST for actions, PUT for updates, DELETE for removal).
- **URL Patterns**: **MUST** follow the URL patterns documented in each method section (e.g., `/v1/message:send`, `/v1/tasks/{id}`).
- **Content-Type**: **MUST** use `application/json` for request and response bodies.
- **HTTP Status Codes**: **MUST** use appropriate HTTP status codes (200, 400, 401, 403, 404, 500, etc.) that correspond to A2A error types.
- **Request/Response Format**: **MUST** use JSON objects that are structurally equivalent to the core A2A data structures.
- **Method Coverage**: **MUST** provide all methods with functionally equivalent behavior to other supported transports.
- **Error Format**: **MUST** return error responses in a consistent JSON format that maps to A2A error types.

#### 3.2.4. Transport Extensions [¶](https://a2a-protocol.org/latest/specification/\#324-transport-extensions "Permanent link")

Additional transport protocols **MAY** be defined as extensions to the core A2A specification. Such extensions:

- **MUST** maintain functional equivalence with the core transports
- **MUST** use clear namespace identifiers to avoid conflicts
- **MUST** be clearly documented and specified
- **SHOULD** provide migration paths from core transports

### 3.3. Streaming Transport (Server-Sent Events) [¶](https://a2a-protocol.org/latest/specification/\#33-streaming-transport-server-sent-events "Permanent link")

Streaming capabilities are **transport-specific**:

#### 3.3.1. JSON-RPC 2.0 Streaming [¶](https://a2a-protocol.org/latest/specification/\#331-json-rpc-20-streaming "Permanent link")

When streaming is used for methods like `message/stream` or `tasks/resubscribe`:

- The server responds with an HTTP `200 OK` status and a `Content-Type` header of `text/event-stream`.
- The body of this HTTP response contains a stream of **[Server-Sent Events (SSE)](https://html.spec.whatwg.org/multipage/server-sent-events.html#server-sent-events)** as defined by the W3C.
- Each SSE `data` field contains a complete JSON-RPC 2.0 Response object (specifically, a [`SendStreamingMessageResponse`](https://a2a-protocol.org/latest/specification/#721-sendstreamingmessageresponse-object)).

#### 3.3.2. gRPC Streaming [¶](https://a2a-protocol.org/latest/specification/\#332-grpc-streaming "Permanent link")

gRPC transport uses **server streaming RPCs** for streaming operations as defined in the Protocol Buffers specification.

#### 3.3.3. HTTP+JSON/REST Streaming [¶](https://a2a-protocol.org/latest/specification/\#333-httpjsonrest-streaming "Permanent link")

If REST transport is supported it **MUST** implement streaming using Server-Sent Events similar to JSON-RPC.

### 3.4. Transport Compliance and Interoperability [¶](https://a2a-protocol.org/latest/specification/\#34-transport-compliance-and-interoperability "Permanent link")

#### 3.4.1. Functional Equivalence Requirements [¶](https://a2a-protocol.org/latest/specification/\#341-functional-equivalence-requirements "Permanent link")

When an agent supports multiple transports, all supported transports **MUST**:

- **Identical Functionality**: Provide the same set of operations and capabilities.
- **Consistent Behavior**: Return semantically equivalent results for the same requests.
- **Same Error Handling**: Map errors consistently across transports using the error codes defined in [Section 8](https://a2a-protocol.org/latest/specification/#8-error-handling).
- **Equivalent Authentication**: Support the same authentication schemes declared in the `AgentCard`.

#### 3.4.2. Transport Selection and Negotiation [¶](https://a2a-protocol.org/latest/specification/\#342-transport-selection-and-negotiation "Permanent link")

- **Agent Declaration**: Agents **MUST** declare all supported transports in their `AgentCard` using the `preferredTransport` and `additionalInterfaces` fields.
- **Client Choice**: Clients **MAY** choose any transport declared by the agent.
- **No Transport Negotiation**: A2A does not define a dynamic transport negotiation protocol. Clients select a transport based on the static `AgentCard` information.
- **Fallback Behavior**: Clients **SHOULD** implement fallback logic to try alternative transports if their preferred transport fails. The specific fallback strategy is implementation-dependent.

#### 3.4.3. Transport-Specific Extensions [¶](https://a2a-protocol.org/latest/specification/\#343-transport-specific-extensions "Permanent link")

Transports **MAY** provide transport-specific optimizations or extensions that do not compromise functional equivalence:

- **gRPC**: May leverage gRPC-specific features like bidirectional streaming, metadata, or custom status codes.
- **REST**: May provide additional HTTP caching headers or support HTTP conditional requests.
- **JSON-RPC**: May include additional fields in the JSON-RPC request/response objects that do not conflict with the core specification.

Such extensions **MUST** be backward-compatible and **MUST NOT** break interoperability with clients that do not support the extensions.

### 3.5. Method Mapping and Naming Conventions [¶](https://a2a-protocol.org/latest/specification/\#35-method-mapping-and-naming-conventions "Permanent link")

To ensure consistency and predictability across different transports, A2A defines normative method mapping rules.

#### 3.5.1. JSON-RPC Method Naming [¶](https://a2a-protocol.org/latest/specification/\#351-json-rpc-method-naming "Permanent link")

JSON-RPC methods **MUST** follow the pattern: `{category}/{action}` where:

- `category` represents the resource type (e.g., "message", "tasks", "agent")
- `action` represents the operation (e.g., "send", "get", "cancel")
- Nested actions use forward slashes (e.g., "tasks/pushNotificationConfig/set")

#### 3.5.2. gRPC Method Naming [¶](https://a2a-protocol.org/latest/specification/\#352-grpc-method-naming "Permanent link")

gRPC methods **MUST** follow Protocol Buffers service conventions using PascalCase:

- Convert JSON-RPC category/action to PascalCase compound words
- Use standard gRPC method prefixes (Get, Set, List, Create, Delete, Cancel)

#### 3.5.3. HTTP+JSON/REST Method Naming [¶](https://a2a-protocol.org/latest/specification/\#353-httpjsonrest-method-naming "Permanent link")

REST endpoints **MUST** follow RESTful URL patterns with appropriate HTTP verbs:

- Use resource-based URLs: `/v1/{resource}[/{id}][:{action}]`
- Use standard HTTP methods aligned with REST semantics
- Use colon notation for non-CRUD actions

#### 3.5.4. Method Mapping Compliance [¶](https://a2a-protocol.org/latest/specification/\#354-method-mapping-compliance "Permanent link")

When implementing multiple transports, agents **MUST**:

- **Use standard mappings**: Follow the method mappings defined in sections 3.5.2 and 3.5.3.
- **Maintain functional equivalence**: Each transport-specific method **MUST** provide identical functionality across all supported transports.
- **Consistent parameters**: Use equivalent parameter structures across transports (accounting for transport-specific serialization differences).
- **Equivalent responses**: Return semantically equivalent responses across all transports for the same operation.

#### 3.5.5. Extension Method Naming [¶](https://a2a-protocol.org/latest/specification/\#355-extension-method-naming "Permanent link")

For custom or extension methods not defined in the core A2A specification:

- **JSON-RPC**: Follow the `{category}/{action}` pattern with a clear namespace (e.g., `myorg.extension/action`)
- **gRPC**: Use appropriate service and method names following Protocol Buffers conventions
- **REST**: Use clear resource-based URLs with appropriate HTTP methods

Extension methods **MUST** be clearly documented and **MUST NOT** conflict with core A2A method names or semantics.

#### 3.5.6. Method Mapping Reference Table [¶](https://a2a-protocol.org/latest/specification/\#356-method-mapping-reference-table "Permanent link")

For quick reference, the following table summarizes the method mappings across all transports:

| JSON-RPC Method | gRPC Method | REST Endpoint | Description |
| --- | --- | --- | --- |
| `message/send` | `SendMessage` | `POST /v1/message:send` | Send message to agent |
| `message/stream` | `SendStreamingMessage` | `POST /v1/message:stream` | Send message with streaming |
| `tasks/get` | `GetTask` | `GET /v1/tasks/{id}` | Get task status |
| `tasks/list` | `ListTask` | `GET /v1/tasks` | List tasks (gRPC/REST only) |
| `tasks/cancel` | `CancelTask` | `POST /v1/tasks/{id}:cancel` | Cancel task |
| `tasks/resubscribe` | `TaskSubscription` | `POST /v1/tasks/{id}:subscribe` | Resume task streaming |
| `tasks/pushNotificationConfig/set` | `CreateTaskPushNotification` | `POST /v1/tasks/{id}/pushNotificationConfigs` | Set push notification config |
| `tasks/pushNotificationConfig/get` | `GetTaskPushNotification` | `GET /v1/tasks/{id}/pushNotificationConfigs/{configId}` | Get push notification config |
| `tasks/pushNotificationConfig/list` | `ListTaskPushNotification` | `GET /v1/tasks/{id}/pushNotificationConfigs` | List push notification configs |
| `tasks/pushNotificationConfig/delete` | `DeleteTaskPushNotification` | `DELETE /v1/tasks/{id}/pushNotificationConfigs/{configId}` | Delete push notification config |
| `agent/authenticatedExtendedCard` | `GetAgentCard` | `GET /v1/card` | Get authenticated agent card |

## 4\. Authentication and Authorization [¶](https://a2a-protocol.org/latest/specification/\#4-authentication-and-authorization "Permanent link")

A2A treats agents as standard enterprise applications, relying on established web security practices. Identity information is **not** transmitted within A2A JSON-RPC payloads; it is handled at the HTTP transport layer.

For a comprehensive guide on enterprise security aspects, see [Enterprise-Ready Features](https://a2a-protocol.org/latest/topics/enterprise-ready/).

### 4.1. Transport Security [¶](https://a2a-protocol.org/latest/specification/\#41-transport-security "Permanent link")

As stated in section 3.1, production deployments **MUST** use HTTPS. Implementations **SHOULD** use modern [TLS](https://datatracker.ietf.org/doc/html/rfc8446) configurations (TLS 1.3+ recommended) with strong cipher suites.

### 4.2. Server Identity Verification [¶](https://a2a-protocol.org/latest/specification/\#42-server-identity-verification "Permanent link")

A2A Clients **SHOULD** verify the A2A Server's identity by validating its TLS certificate against trusted certificate authorities (CAs) during the TLS handshake.

### 4.3. Client/User Identity & Authentication Process [¶](https://a2a-protocol.org/latest/specification/\#43-clientuser-identity-authentication-process "Permanent link")

1. **Discovery of Requirements:** The client discovers the server's required authentication schemes via the `authentication` field in the [`AgentCard`](https://a2a-protocol.org/latest/specification/#55-agentcard-object-structure). Scheme names often align with [OpenAPI Authentication methods](https://swagger.io/docs/specification/authentication/) (e.g., "Bearer" for OAuth 2.0 tokens, "Basic" for Basic Auth, "ApiKey" for API keys).
2. **Credential Acquisition (Out-of-Band):** The client obtains the necessary credentials (e.g., API keys, OAuth tokens, JWTs) through an **out-of-band process** specific to the required authentication scheme and the identity provider. This process is outside the scope of the A2A protocol itself.
3. **Credential Transmission:** The client includes these credentials in the appropriate [HTTP headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers) (e.g., `Authorization: Bearer <token>`, `X-API-Key: <value>`) of every A2A request sent to the server.

### 4.4. Server Responsibilities for Authentication [¶](https://a2a-protocol.org/latest/specification/\#44-server-responsibilities-for-authentication "Permanent link")

The A2A Server:

- **MUST** authenticate every incoming request based on the provided HTTP credentials and its declared authentication requirements from its Agent Card.
- **SHOULD** use standard HTTP status codes like [`401 Unauthorized`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/401) or [`403 Forbidden`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/403) for authentication challenges or rejections.
- **SHOULD** include relevant HTTP headers (e.g., [`WWW-Authenticate`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate)) with `401 Unauthorized` responses to indicate the required authentication scheme(s), guiding the client.

### 4.5. In-Task Authentication (Secondary Credentials) [¶](https://a2a-protocol.org/latest/specification/\#45-in-task-authentication-secondary-credentials "Permanent link")

If an agent, during the execution of a task, requires _additional_ credentials for a _different_ system or resource (e.g., to access a specific tool on behalf of the user that requires its own auth):

1. It **SHOULD** transition the A2A task to the `auth-required` state (see [`TaskState`](https://a2a-protocol.org/latest/specification/#63-taskstate-enum)).
2. The accompanying `TaskStatus.message` (often a [`DataPart`](https://a2a-protocol.org/latest/specification/#653-datapart-object)) **SHOULD** provide details about the required secondary authentication, potentially using an [`PushNotificationAuthenticationInfo`](https://a2a-protocol.org/latest/specification/#69-pushnotificationauthenticationinfo-object)-like structure to describe the need.
3. The A2A Client then obtains these new credentials out-of-band and provides them in a subsequent [`message/send`](https://a2a-protocol.org/latest/specification/#71-messagesend) or [`message/stream`](https://a2a-protocol.org/latest/specification/#72-messagestream) request. How these credentials are used (e.g., passed as data within the A2A message if the agent is proxying, or used by the client to interact directly with the secondary system) depends on the specific scenario.

### 4.6. Authorization [¶](https://a2a-protocol.org/latest/specification/\#46-authorization "Permanent link")

Once a client is authenticated, the A2A Server is responsible for authorizing the request based on the authenticated client/user identity and its own policies. Authorization logic is implementation-specific and MAY be enforced based on:

- The specific skills requested (e.g., as identified by `AgentSkill.id` from the Agent Card).
- The actions attempted within the task.
- Data access policies relevant to the resources the agent manages.
- OAuth scopes associated with the presented token, if applicable.

Servers should implement the principle of least privilege.

## 5\. Agent Discovery: The Agent Card [¶](https://a2a-protocol.org/latest/specification/\#5-agent-discovery-the-agent-card "Permanent link")

### 5.1. Purpose [¶](https://a2a-protocol.org/latest/specification/\#51-purpose "Permanent link")

A2A Servers **MUST** make an Agent Card available. The Agent Card is a JSON document that describes the server's identity, capabilities, skills, service endpoint URL, and how clients should authenticate and interact with it. Clients use this information for discovering suitable agents and for configuring their interactions.

For more on discovery strategies, see the [Agent Discovery guide](https://a2a-protocol.org/latest/topics/agent-discovery/).

### 5.2. Discovery Mechanisms [¶](https://a2a-protocol.org/latest/specification/\#52-discovery-mechanisms "Permanent link")

Clients can find Agent Cards through various methods, including but not limited to:

- **Well-Known URI:** Accessing a predefined path on the agent's domain (see [Section 5.3](https://a2a-protocol.org/latest/specification/#53-recommended-location)).
- **Registries/Catalogs:** Querying curated catalogs or registries of agents (which might be enterprise-specific, public, or domain-specific).
- **Direct Configuration:** Clients may be pre-configured with the Agent Card URL or the card content itself.

### 5.3. Recommended Location [¶](https://a2a-protocol.org/latest/specification/\#53-recommended-location "Permanent link")

If using the well-known URI strategy, the recommended location for an agent's Agent Card is:
`https://{server_domain}/.well-known/agent-card.json`
This follows the principles of [RFC 8615](https://datatracker.ietf.org/doc/html/rfc8615) for well-known URIs.

### 5.4. Security of Agent Cards [¶](https://a2a-protocol.org/latest/specification/\#54-security-of-agent-cards "Permanent link")

Agent Cards themselves might contain information that is considered sensitive.

- If an Agent Card contains sensitive information, the endpoint serving the card **MUST** be protected by appropriate access controls (e.g., mTLS, network restrictions, authentication required to fetch the card).
- It is generally **NOT RECOMMENDED** to include plaintext secrets (like static API keys) directly in an Agent Card. Prefer authentication schemes where clients obtain dynamic credentials out-of-band.

### 5.5. `AgentCard` Object Structure [¶](https://a2a-protocol.org/latest/specification/\#55-agentcard-object-structure "Permanent link")

```md-code__content
/**
 * The AgentCard is a self-describing manifest for an agent. It provides essential
 * metadata including the agent's identity, capabilities, skills, supported
 * communication methods, and security requirements.
 */
export interface AgentCard {
  /**
   * The version of the A2A protocol this agent supports.
   * @default "0.2.6"
   */
  protocolVersion: string;
  /**
   * A human-readable name for the agent.
   *
   * @TJS-examples ["Recipe Agent"]
   */
  name: string;
  /**
   * A human-readable description of the agent, assisting users and other agents
   * in understanding its purpose.
   *
   * @TJS-examples ["Agent that helps users with recipes and cooking."]
   */
  description: string;
  /**
   * The preferred endpoint URL for interacting with the agent.
   * This URL MUST support the transport specified by 'preferredTransport'.
   *
   * @TJS-examples ["https://api.example.com/a2a/v1"]
   */
  url: string;
  /**
   * The transport protocol for the preferred endpoint (the main 'url' field).
   * If not specified, defaults to 'JSONRPC'.
   *
   * IMPORTANT: The transport specified here MUST be available at the main 'url'.
   * This creates a binding between the main URL and its supported transport protocol.
   * Clients should prefer this transport and URL combination when both are supported.
   *
   * @default "JSONRPC"
   * @TJS-examples ["JSONRPC", "GRPC", "HTTP+JSON"]
   */
  preferredTransport?: TransportProtocol | string;
  /**
   * A list of additional supported interfaces (transport and URL combinations).
   * This allows agents to expose multiple transports, potentially at different URLs.
   *
   * Best practices:
   * - SHOULD include all supported transports for completeness
   * - SHOULD include an entry matching the main 'url' and 'preferredTransport'
   * - MAY reuse URLs if multiple transports are available at the same endpoint
   * - MUST accurately declare the transport available at each URL
   *
   * Clients can select any interface from this list based on their transport capabilities
   * and preferences. This enables transport negotiation and fallback scenarios.
   */
  additionalInterfaces?: AgentInterface[];
  /** An optional URL to an icon for the agent. */
  iconUrl?: string;
  /** Information about the agent's service provider. */
  provider?: AgentProvider;
  /**
   * The agent's own version number. The format is defined by the provider.
   *
   * @TJS-examples ["1.0.0"]
   */
  version: string;
  /** An optional URL to the agent's documentation. */
  documentationUrl?: string;
  /** A declaration of optional capabilities supported by the agent. */
  capabilities: AgentCapabilities;
  /**
   * A declaration of the security schemes available to authorize requests. The key is the
   * scheme name. Follows the OpenAPI 3.0 Security Scheme Object.
   */
  securitySchemes?: { [scheme: string]: SecurityScheme };
  /**
   * A list of security requirement objects that apply to all agent interactions. Each object
   * lists security schemes that can be used. Follows the OpenAPI 3.0 Security Requirement Object.
   */
  security?: { [scheme: string]: string[] }[];
  /**
   * Default set of supported input MIME types for all skills, which can be
   * overridden on a per-skill basis.
   */
  defaultInputModes: string[];
  /**
   * Default set of supported output MIME types for all skills, which can be
   * overridden on a per-skill basis.
   */
  defaultOutputModes: string[];
  /** The set of skills, or distinct capabilities, that the agent can perform. */
  skills: AgentSkill[];
  /**
   * If true, the agent can provide an extended agent card with additional details
   * to authenticated users. Defaults to false.
   */
  supportsAuthenticatedExtendedCard?: boolean;
  /** JSON Web Signatures computed for this AgentCard. */
  signatures?: AgentCardSignature[];
}

```

#### 5.5.1. `AgentProvider` Object [¶](https://a2a-protocol.org/latest/specification/\#551-agentprovider-object "Permanent link")

Information about the organization or entity providing the agent.

```md-code__content
/**
 * Represents the service provider of an agent.
 *
 * @TJS-examples [{ "organization": "Google", "url": "https://ai.google.dev" }]
 */
export interface AgentProvider {
  /** The name of the agent provider's organization. */
  organization: string;
  /** A URL for the agent provider's website or relevant documentation. */
  url: string;
}

```

#### 5.5.2. `AgentCapabilities` Object [¶](https://a2a-protocol.org/latest/specification/\#552-agentcapabilities-object "Permanent link")

Specifies optional A2A protocol features supported by the agent.

```md-code__content
/**
 * Defines optional capabilities supported by an agent.
 */
export interface AgentCapabilities {
  /** Indicates if the agent supports Server-Sent Events (SSE) for streaming responses. */
  streaming?: boolean;
  /** Indicates if the agent supports sending push notifications for asynchronous task updates. */
  pushNotifications?: boolean;
  /** Indicates if the agent provides a history of state transitions for a task. */
  stateTransitionHistory?: boolean;
  /** A list of protocol extensions supported by the agent. */
  extensions?: AgentExtension[];
}

```

#### 5.5.2.1. `AgentExtension` Object [¶](https://a2a-protocol.org/latest/specification/\#5521-agentextension-object "Permanent link")

Specifies an extension to the A2A protocol supported by the agent.

```md-code__content
/**
 * A declaration of a protocol extension supported by an Agent.
 *
 * @TJS-examples [{"uri": "https://developers.google.com/identity/protocols/oauth2", "description": "Google OAuth 2.0 authentication", "required": false}]
 */
export interface AgentExtension {
  /** The unique URI identifying the extension. */
  uri: string;
  /** A human-readable description of how this agent uses the extension. */
  description?: string;
  /**
   * If true, the client must understand and comply with the extension's requirements
   * to interact with the agent.
   */
  required?: boolean;
  /** Optional, extension-specific configuration parameters. */
  params?: { [key: string]: any };
}

```

#### 5.5.3. `SecurityScheme` Object [¶](https://a2a-protocol.org/latest/specification/\#553-securityscheme-object "Permanent link")

Describes the authentication requirements for accessing the agent's `url` endpoint. Refer [Sample Agent Card](https://a2a-protocol.org/latest/specification/#57-sample-agent-card) for an example.

```md-code__content
/**
 * Defines a security scheme that can be used to secure an agent's endpoints.
 * This is a discriminated union type based on the OpenAPI 3.0 Security Scheme Object.
 *
 * @see {@link https://swagger.io/specification/#security-scheme-object}
 */
export type SecurityScheme =
  | APIKeySecurityScheme
  | HTTPAuthSecurityScheme
  | OAuth2SecurityScheme
  | OpenIdConnectSecurityScheme;

```

#### 5.5.4. `AgentSkill` Object [¶](https://a2a-protocol.org/latest/specification/\#554-agentskill-object "Permanent link")

Describes a specific capability, function, or area of expertise the agent can perform or address.

```md-code__content
/**
 * Represents a distinct capability or function that an agent can perform.
 */
export interface AgentSkill {
  /** A unique identifier for the agent's skill. */
  id: string;
  /** A human-readable name for the skill. */
  name: string;
  /**
   * A detailed description of the skill, intended to help clients or users
   * understand its purpose and functionality.
   */
  description: string;
  /**
   * A set of keywords describing the skill's capabilities.
   *
   * @TJS-examples [["cooking", "customer support", "billing"]]
   */
  tags: string[];
  /**
   * Example prompts or scenarios that this skill can handle. Provides a hint to
   * the client on how to use the skill.
   *
   * @TJS-examples [["I need a recipe for bread"]]
   */
  examples?: string[];
  /**
   * The set of supported input MIME types for this skill, overriding the agent's defaults.
   */
  inputModes?: string[];
  /**
   * The set of supported output MIME types for this skill, overriding the agent's defaults.
   */
  outputModes?: string[];
}

```

#### 5.5.5. `AgentInterface` Object [¶](https://a2a-protocol.org/latest/specification/\#555-agentinterface-object "Permanent link")

Provides a declaration of a combination of the target URL and the supported transport to interact with the agent. This enables agents to expose the same functionality through multiple transport protocols.

```md-code__content
/**
 * Supported A2A transport protocols.
 */
export enum TransportProtocol {
  JSONRPC = "JSONRPC", // JSON-RPC 2.0 over HTTP (mandatory)
  GRPC = "GRPC", // gRPC over HTTP/2 (optional)
  HTTP_JSON = "HTTP+JSON", // REST-style HTTP with JSON (optional)
}

```

```md-code__content
/**
 * Declares a combination of a target URL and a transport protocol for interacting with the agent.
 * This allows agents to expose the same functionality over multiple transport mechanisms.
 */
export interface AgentInterface {
  /**
   * The URL where this interface is available. Must be a valid absolute HTTPS URL in production.
   * @TJS-examples ["https://api.example.com/a2a/v1", "https://grpc.example.com/a2a", "https://rest.example.com/v1"]
   */
  url: string;
  /**
   * The transport protocol supported at this URL.
   *
   * @TJS-examples ["JSONRPC", "GRPC", "HTTP+JSON"]
   */
  transport: TransportProtocol | string;
}

```

The `transport` field **SHOULD** use one of the core A2A transport protocol values:

- `"JSONRPC"`: JSON-RPC 2.0 over HTTP
- `"GRPC"`: gRPC over HTTP/2
- `"HTTP+JSON"`: REST-style HTTP with JSON

Additional transport values **MAY** be used for future extensions, but such extensions **MUST** not conflict with core A2A protocol functionality.

#### 5.5.6. `AgentCardSignature` Object [¶](https://a2a-protocol.org/latest/specification/\#556-agentcardsignature-object "Permanent link")

Represents a JSON Web Signature (JWS) used to verify the integrity of the AgentCard.

```md-code__content
/**
 * AgentCardSignature represents a JWS signature of an AgentCard.
 * This follows the JSON format of an RFC 7515 JSON Web Signature (JWS).
 */
export interface AgentCardSignature {
  /**
   * The protected JWS header for the signature. This is a Base64url-encoded
   * JSON object, as per RFC 7515.
   */
  protected: string;
  /** The computed signature, Base64url-encoded. */
  signature: string;
  /** The unprotected JWS header values. */
  header?: { [key: string]: any };
}

```

### 5.6. Transport Declaration and URL Relationships [¶](https://a2a-protocol.org/latest/specification/\#56-transport-declaration-and-url-relationships "Permanent link")

The AgentCard **MUST** properly declare the relationship between URLs and transport protocols:

#### 5.6.1. Main URL and Preferred Transport [¶](https://a2a-protocol.org/latest/specification/\#561-main-url-and-preferred-transport "Permanent link")

- **Main URL requirement**: The `url` field **MUST** specify the primary endpoint for the agent.
- **Transport correspondence**: The transport protocol available at the main `url` **MUST** match the `preferredTransport` field.
- **Required declaration**: The `preferredTransport` field is **REQUIRED** and **MUST** be present in every `AgentCard`.
- **Transport availability**: The main `url` **MUST** support the transport protocol declared in `preferredTransport`.

#### 5.6.2. Additional Interfaces [¶](https://a2a-protocol.org/latest/specification/\#562-additional-interfaces "Permanent link")

- **URL uniqueness**: Each `AgentInterface` in `additionalInterfaces` **SHOULD** specify a distinct URL for clarity, but **MAY** reuse URLs if multiple transport protocols are available at the same endpoint.
- **Transport declaration**: Each `AgentInterface` **MUST** accurately declare the transport protocol available at its specified URL.
- **Completeness**: The `additionalInterfaces` array **SHOULD** include all supported transports, including the main URL's transport for completeness.

#### 5.6.3. Client Transport Selection Rules [¶](https://a2a-protocol.org/latest/specification/\#563-client-transport-selection-rules "Permanent link")

Clients **MUST** follow these rules when selecting a transport:

1. **Parse transport declarations**: Extract available transports from both the main `url`/ `preferredTransport` combination and all `additionalInterfaces`.
2. **Prefer declared preference**: If the client supports the `preferredTransport`, it **SHOULD** use the main `url`.
3. **Fallback selection**: If the preferred transport is not supported by the client, it **MAY** select any supported transport from `additionalInterfaces`.
4. **Graceful degradation**: Clients **SHOULD** implement fallback logic to try alternative transports if their first choice fails.
5. **URL-transport matching**: Clients **MUST** use the correct URL for the selected transport protocol as declared in the AgentCard.

#### 5.6.4. Validation Requirements [¶](https://a2a-protocol.org/latest/specification/\#564-validation-requirements "Permanent link")

Agent Cards **MUST** satisfy these validation requirements:

- **Transport consistency**: The `preferredTransport` value **MUST** be present and **MUST** be available at the main `url`.
- **Interface completeness**: If `additionalInterfaces` is provided, it **SHOULD** include an entry corresponding to the main `url` and `preferredTransport`.
- **No conflicts**: The same URL **MUST NOT** declare conflicting transport protocols across different interface declarations.
- **Minimum transport requirement**: The agent **MUST** declare at least one supported transport protocol through either the main `url`/ `preferredTransport` combination or `additionalInterfaces`.

### 5.7. Sample Agent Card [¶](https://a2a-protocol.org/latest/specification/\#57-sample-agent-card "Permanent link")

```md-code__content
{
  "protocolVersion": "0.2.9",
  "name": "GeoSpatial Route Planner Agent",
  "description": "Provides advanced route planning, traffic analysis, and custom map generation services. This agent can calculate optimal routes, estimate travel times considering real-time traffic, and create personalized maps with points of interest.",
  "url": "https://georoute-agent.example.com/a2a/v1",
  "preferredTransport": "JSONRPC",
  "additionalInterfaces" : [\
    {"url": "https://georoute-agent.example.com/a2a/v1", "transport": "JSONRPC"},\
    {"url": "https://georoute-agent.example.com/a2a/grpc", "transport": "GRPC"},\
    {"url": "https://georoute-agent.example.com/a2a/json", "transport": "HTTP+JSON"}\
  ],
  "provider": {
    "organization": "Example Geo Services Inc.",
    "url": "https://www.examplegeoservices.com"
  },
  "iconUrl": "https://georoute-agent.example.com/icon.png",
  "version": "1.2.0",
  "documentationUrl": "https://docs.examplegeoservices.com/georoute-agent/api",
  "capabilities": {
    "streaming": true,
    "pushNotifications": true,
    "stateTransitionHistory": false
  },
  "securitySchemes": {
    "google": {
      "type": "openIdConnect",
      "openIdConnectUrl": "https://accounts.google.com/.well-known/openid-configuration"
    }
  },
  "security": [{ "google": ["openid", "profile", "email"] }],
  "defaultInputModes": ["application/json", "text/plain"],
  "defaultOutputModes": ["application/json", "image/png"],
  "skills": [\
    {\
      "id": "route-optimizer-traffic",\
      "name": "Traffic-Aware Route Optimizer",\
      "description": "Calculates the optimal driving route between two or more locations, taking into account real-time traffic conditions, road closures, and user preferences (e.g., avoid tolls, prefer highways).",\
      "tags": ["maps", "routing", "navigation", "directions", "traffic"],\
      "examples": [\
        "Plan a route from '1600 Amphitheatre Parkway, Mountain View, CA' to 'San Francisco International Airport' avoiding tolls.",\
        "{\"origin\": {\"lat\": 37.422, \"lng\": -122.084}, \"destination\": {\"lat\": 37.7749, \"lng\": -122.4194}, \"preferences\": [\"avoid_ferries\"]}"\
      ],\
      "inputModes": ["application/json", "text/plain"],\
      "outputModes": [\
        "application/json",\
        "application/vnd.geo+json",\
        "text/html"\
      ]\
    },\
    {\
      "id": "custom-map-generator",\
      "name": "Personalized Map Generator",\
      "description": "Creates custom map images or interactive map views based on user-defined points of interest, routes, and style preferences. Can overlay data layers.",\
      "tags": ["maps", "customization", "visualization", "cartography"],\
      "examples": [\
        "Generate a map of my upcoming road trip with all planned stops highlighted.",\
        "Show me a map visualizing all coffee shops within a 1-mile radius of my current location."\
      ],\
      "inputModes": ["application/json"],\
      "outputModes": [\
        "image/png",\
        "image/jpeg",\
        "application/json",\
        "text/html"\
      ]\
    }\
  ],
  "supportsAuthenticatedExtendedCard": true,
  "signatures": [\
    {\
      "protected": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpPU0UiLCJraWQiOiJrZXktMSIsImprdSI6Imh0dHBzOi8vZXhhbXBsZS5jb20vYWdlbnQvandrcy5qc29uIn0",\
      "signature": "QFdkNLNszlGj3z3u0YQGt_T9LixY3qtdQpZmsTdDHDe3fXV9y9-B3m2-XgCpzuhiLt8E0tV6HXoZKHv4GtHgKQ"\
    }\
  ]
}

```

## 6\. Protocol Data Objects [¶](https://a2a-protocol.org/latest/specification/\#6-protocol-data-objects "Permanent link")

These objects define the structure of data exchanged within the JSON-RPC methods of the A2A protocol.

### 6.1. `Task` Object [¶](https://a2a-protocol.org/latest/specification/\#61-task-object "Permanent link")

Represents the stateful unit of work being processed by the A2A Server for an A2A Client. A task encapsulates the entire interaction related to a specific goal or request. A task which has reached a terminal state (completed, canceled, rejected, or failed) can't be restarted. Tasks in completed state SHOULD use artifacts for returning the generated output to the clients. For more information, refer to the [Life of a Task guide](https://a2a-protocol.org/latest/topics/life-of-a-task/).

```md-code__content
/**
 * Represents a single, stateful operation or conversation between a client and an agent.
 */
export interface Task {
  /** A unique identifier for the task, generated by the server for a new task. */
  id: string;
  /** A server-generated identifier for maintaining context across multiple related tasks or interactions. */
  contextId: string;
  /** The current status of the task, including its state and a descriptive message. */
  status: TaskStatus;
  /** An array of messages exchanged during the task, representing the conversation history. */
  history?: Message[];
  /** A collection of artifacts generated by the agent during the execution of the task. */
  artifacts?: Artifact[];
  /** Optional metadata for extensions. The key is an extension-specific identifier. */
  metadata?: {
    [key: string]: any;
  };
  /** The type of this object, used as a discriminator. Always 'task' for a Task. */
  readonly kind: "task";
}

```

### 6.2. `TaskStatus` Object [¶](https://a2a-protocol.org/latest/specification/\#62-taskstatus-object "Permanent link")

Represents the current state and associated context (e.g., a message from the agent) of a `Task`.

```md-code__content
/**
 * Represents the status of a task at a specific point in time.
 */
export interface TaskStatus {
  /** The current state of the task's lifecycle. */
  state: TaskState;
  /** An optional, human-readable message providing more details about the current status. */
  message?: Message;
  /**
   * An ISO 8601 datetime string indicating when this status was recorded.
   *
   * @TJS-examples ["2023-10-27T10:00:00Z"]
   */
  timestamp?: string;
}

```

### 6.3. `TaskState` Enum [¶](https://a2a-protocol.org/latest/specification/\#63-taskstate-enum "Permanent link")

Defines the possible lifecycle states of a `Task`.

```md-code__content
/**
 * Defines the lifecycle states of a Task.
 */
export enum TaskState {
  /** The task has been submitted and is awaiting execution. */
  Submitted = "submitted",
  /** The agent is actively working on the task. */
  Working = "working",
  /** The task is paused and waiting for input from the user. */
  InputRequired = "input-required",
  /** The task has been successfully completed. */
  Completed = "completed",
  /** The task has been canceled by the user. */
  Canceled = "canceled",
  /** The task failed due to an error during execution. */
  Failed = "failed",
  /** The task was rejected by the agent and was not started. */
  Rejected = "rejected",
  /** The task requires authentication to proceed. */
  AuthRequired = "auth-required",
  /** The task is in an unknown or indeterminate state. */
  Unknown = "unknown",
}

```

### 6.4. `Message` Object [¶](https://a2a-protocol.org/latest/specification/\#64-message-object "Permanent link")

Represents a single communication turn or a piece of contextual information between a client and an agent. Messages are used for instructions, prompts, replies, and status updates.

```md-code__content
/**
 * Represents a single message in the conversation between a user and an agent.
 */
export interface Message {
  /** Identifies the sender of the message. `user` for the client, `agent` for the service. */
  readonly role: "user" | "agent";
  /**
   * An array of content parts that form the message body. A message can be
   * composed of multiple parts of different types (e.g., text and files).
   */
  parts: Part[];
  /** Optional metadata for extensions. The key is an extension-specific identifier. */
  metadata?: {
    [key: string]: any;
  };
  /** The URIs of extensions that are relevant to this message. */
  extensions?: string[];
  /** A list of other task IDs that this message references for additional context. */
  referenceTaskIds?: string[];
  /** A unique identifier for the message, typically a UUID, generated by the sender. */
  messageId: string;
  /** The identifier of the task this message is part of. Can be omitted for the first message of a new task. */
  taskId?: string;
  /** The context identifier for this message, used to group related interactions. */
  contextId?: string;
  /** The type of this object, used as a discriminator. Always 'message' for a Message. */
  readonly kind: "message";
}

```

### 6.5. `Part` Union Type [¶](https://a2a-protocol.org/latest/specification/\#65-part-union-type "Permanent link")

Represents a distinct piece of content within a `Message` or `Artifact`. A `Part` is a union type representing exportable content as either `TextPart`, `FilePart`, or `DataPart`. All `Part` types also include an optional `metadata` field ( `Record<string, any>`) for part-specific metadata.

```md-code__content
/**
 * A discriminated union representing a part of a message or artifact, which can
 * be text, a file, or structured data.
 */
export type Part = TextPart | FilePart | DataPart;

```

```md-code__content
/**
 * Defines base properties common to all message or artifact parts.
 */
export interface PartBase {
  /** Optional metadata associated with this part. */
  metadata?: {
    [key: string]: any;
  };
}

```

It **MUST** be one of the following:

#### 6.5.1. `TextPart` Object [¶](https://a2a-protocol.org/latest/specification/\#651-textpart-object "Permanent link")

For conveying plain textual content.

```md-code__content
/**
 * Represents a text segment within a message or artifact.
 */
export interface TextPart extends PartBase {
  /** The type of this part, used as a discriminator. Always 'text'. */
  readonly kind: "text";
  /** The string content of the text part. */
  text: string;
}

```

#### 6.5.2. `FilePart` Object [¶](https://a2a-protocol.org/latest/specification/\#652-filepart-object "Permanent link")

For conveying file-based content.

```md-code__content
/**
 * Represents a file segment within a message or artifact. The file content can be
 * provided either directly as bytes or as a URI.
 */
export interface FilePart extends PartBase {
  /** The type of this part, used as a discriminator. Always 'file'. */
  readonly kind: "file";
  /** The file content, represented as either a URI or as base64-encoded bytes. */
  file: FileWithBytes | FileWithUri;
}

```

#### 6.5.3. `DataPart` Object [¶](https://a2a-protocol.org/latest/specification/\#653-datapart-object "Permanent link")

For conveying structured JSON data. Useful for forms, parameters, or any machine-readable information.

```md-code__content
/**
 * Represents a structured data segment (e.g., JSON) within a message or artifact.
 */
export interface DataPart extends PartBase {
  /** The type of this part, used as a discriminator. Always 'data'. */
  readonly kind: "data";
  /** The structured data content. */
  data: {
    [key: string]: any;
  };
}

```

### 6.6 `FileBase` Object [¶](https://a2a-protocol.org/latest/specification/\#66-filebase-object "Permanent link")

Base entity for File Contents.

```md-code__content
/**
 * Defines base properties for a file.
 */
export interface FileBase {
  /** An optional name for the file (e.g., "document.pdf"). */
  name?: string;
  /** The MIME type of the file (e.g., "application/pdf"). */
  mimeType?: string;
}

```

#### 6.6.1 `FileWithBytes` Object [¶](https://a2a-protocol.org/latest/specification/\#661-filewithbytes-object "Permanent link")

Represents the data for a file, used within a `FilePart`.

```md-code__content
/**
 * Represents a file with its content provided directly as a base64-encoded string.
 */
export interface FileWithBytes extends FileBase {
  /** The base64-encoded content of the file. */
  bytes: string;
  /** The `uri` property must be absent when `bytes` is present. */
  uri?: never;
}

```

#### 6.6.2 `FileWithUri` Object [¶](https://a2a-protocol.org/latest/specification/\#662-filewithuri-object "Permanent link")

Represents the URI for a file, used within a `FilePart`.

```md-code__content
/**
 * Represents a file with its content located at a specific URI.
 */
export interface FileWithUri extends FileBase {
  /** A URL pointing to the file's content. */
  uri: string;
  /** The `bytes` property must be absent when `uri` is present. */
  bytes?: never;
}

```

### 6.7. `Artifact` Object [¶](https://a2a-protocol.org/latest/specification/\#67-artifact-object "Permanent link")

Represents a tangible output generated by the agent during a task. Artifacts are the results or products of the agent's work.

```md-code__content
/**
 * Represents a file, data structure, or other resource generated by an agent during a task.
 */
export interface Artifact {
  /** A unique identifier for the artifact within the scope of the task. */
  artifactId: string;
  /** An optional, human-readable name for the artifact. */
  name?: string;
  /** An optional, human-readable description of the artifact. */
  description?: string;
  /** An array of content parts that make up the artifact. */
  parts: Part[];
  /** Optional metadata for extensions. The key is an extension-specific identifier. */
  metadata?: {
    [key: string]: any;
  };
  /** The URIs of extensions that are relevant to this artifact. */
  extensions?: string[];
}

```

### 6.8. `PushNotificationConfig` Object [¶](https://a2a-protocol.org/latest/specification/\#68-pushnotificationconfig-object "Permanent link")

Configuration provided by the client to the server for sending asynchronous push notifications about task updates.

```md-code__content
/**
 * Defines the configuration for setting up push notifications for task updates.
 */
export interface PushNotificationConfig {
  /**
   * A unique ID for the push notification configuration, created by the server
   * to support multiple notification callbacks.
   */
  id?: string;
  /** The callback URL where the agent should send push notifications. */
  url: string;
  /** A unique token for this task or session to validate incoming push notifications. */
  token?: string;
  /** Optional authentication details for the agent to use when calling the notification URL. */
  authentication?: PushNotificationAuthenticationInfo;
}

```

### 6.9. `PushNotificationAuthenticationInfo` Object [¶](https://a2a-protocol.org/latest/specification/\#69-pushnotificationauthenticationinfo-object "Permanent link")

A generic structure for specifying authentication requirements, typically used within `PushNotificationConfig` to describe how the A2A Server should authenticate to the client's webhook.

```md-code__content
/**
 * Defines authentication details for a push notification endpoint.
 */
export interface PushNotificationAuthenticationInfo {
  /** A list of supported authentication schemes (e.g., 'Basic', 'Bearer'). */
  schemes: string[];
  /** Optional credentials required by the push notification endpoint. */
  credentials?: string;
}

```

### 6.10. `TaskPushNotificationConfig` Object [¶](https://a2a-protocol.org/latest/specification/\#610-taskpushnotificationconfig-object "Permanent link")

Used as the `params` object for the [`tasks/pushNotificationConfig/set`](https://a2a-protocol.org/latest/specification/#75-taskspushnotificationconfigset) method and as the `result` object for the [`tasks/pushNotificationConfig/get`](https://a2a-protocol.org/latest/specification/#76-taskspushnotificationconfigget) method.

```md-code__content
/**
 * A container associating a push notification configuration with a specific task.
 */
export interface TaskPushNotificationConfig {
  /** The ID of the task. */
  taskId: string;
  /** The push notification configuration for this task. */
  pushNotificationConfig: PushNotificationConfig;
}

```

### 6.11. JSON-RPC Structures [¶](https://a2a-protocol.org/latest/specification/\#611-json-rpc-structures "Permanent link")

A2A adheres to the standard [JSON-RPC 2.0](https://www.jsonrpc.org/specification) structures for requests and responses.

#### 6.11.1. `JSONRPCRequest` Object [¶](https://a2a-protocol.org/latest/specification/\#6111-jsonrpcrequest-object "Permanent link")

All A2A method calls are encapsulated in a JSON-RPC Request object.

- `jsonrpc`: A String specifying the version of the JSON-RPC protocol. **MUST** be exactly `"2.0"`.
- `method`: A String containing the name of the method to be invoked (e.g., `"message/send"`, `"tasks/get"`).
- `params`: A Structured value that holds the parameter values to be used during the invocation of the method. This member **MAY** be omitted if the method expects no parameters. A2A methods typically use an `object` for `params`.
- `id`: An identifier established by the Client that **MUST** contain a String, Number, or `NULL` value if included. If it is not included it is assumed to be a notification. The value **SHOULD NOT** be `NULL` for requests expecting a response, and Numbers **SHOULD NOT** contain fractional parts. The Server **MUST** reply with the same value in the Response object if included. This member is used to correlate the context between the two objects. A2A methods typically expect a response or stream, so `id` will usually be present and non-null.

#### 6.11.2. `JSONRPCResponse` Object [¶](https://a2a-protocol.org/latest/specification/\#6112-jsonrpcresponse-object "Permanent link")

Responses from the A2A Server are encapsulated in a JSON-RPC Response object.

- `jsonrpc`: A String specifying the version of the JSON-RPC protocol. **MUST** be exactly `"2.0"`.
- `id`: This member is **REQUIRED**. It **MUST** be the same as the value of the `id` member in the Request Object. If there was an error in detecting the `id` in the Request object (e.g. Parse error/Invalid Request), it **MUST** be `null`.
- **EITHER** `result`: This member is **REQUIRED** on success. This member **MUST NOT** exist if there was an error invoking the method. The value of this member is determined by the method invoked on the Server.
- **OR** `error`: This member is **REQUIRED** on failure. This member **MUST NOT** exist if there was no error triggered during invocation. The value of this member **MUST** be an [`JSONRPCError`](https://a2a-protocol.org/latest/specification/#612-jsonrpcerror-object) object.
- The members `result` and `error` are mutually exclusive: one **MUST** be present, and the other **MUST NOT**.

### 6.12. `JSONRPCError` Object [¶](https://a2a-protocol.org/latest/specification/\#612-jsonrpcerror-object "Permanent link")

When a JSON-RPC call encounters an error, the Response Object will contain an `error` member with a value of this structure.

```md-code__content
/**
 * Represents a JSON-RPC 2.0 Error object, included in an error response.
 */
export interface JSONRPCError {
  /**
   * A number that indicates the error type that occurred.
   */
  code: number;
  /**
   * A string providing a short description of the error.
   */
  message: string;
  /**
   * A primitive or structured value containing additional information about the error.
   * This may be omitted.
   */
  data?: any;
}

```

## 7\. Protocol RPC Methods [¶](https://a2a-protocol.org/latest/specification/\#7-protocol-rpc-methods "Permanent link")

All A2A RPC methods are invoked by the A2A Client by sending an HTTP POST request to the A2A Server's `url` (as specified in its `AgentCard`). The body of the HTTP POST request **MUST** be a `JSONRPCRequest` object, and the `Content-Type` header **MUST** be `application/json`.

The A2A Server's HTTP response body **MUST** be a `JSONRPCResponse` object (or, for streaming methods, an SSE stream where each event's data is a `JSONRPCResponse`). The `Content-Type` for JSON-RPC responses is `application/json`. For SSE streams, it is `text/event-stream`.

### 7.1. `message/send` [¶](https://a2a-protocol.org/latest/specification/\#71-messagesend "Permanent link")

Sends a message to an agent to initiate a new interaction or to continue an existing one. This method is suitable for synchronous request/response interactions or when client-side polling (using `tasks/get`) is acceptable for monitoring longer-running tasks. A task which has reached a terminal state (completed, canceled, rejected, or failed) can't be restarted. Sending a message to such a task will result in an error. For more information, refer to the [Life of a Task guide](https://a2a-protocol.org/latest/topics/life-of-a-task/).

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc)[gRPC](https://a2a-protocol.org/latest/specification/#grpc)[REST](https://a2a-protocol.org/latest/specification/#rest)

- **URL:** `message/send`
- **HTTP Method:** `POST`
- **Payload**: [`MessageSendParams`](https://a2a-protocol.org/latest/specification/#711-messagesendparams-object)
- **Response**: [`Task` \| `Message`](https://a2a-protocol.org/latest/specification/#61-task-object) (A message object or the current or final state of the task after processing the message).

- **URL:** `SendMessage`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message SendMessageRequest {
    Message msg = 1;
    SendMessageConfiguration configuration = 2;
}

```

- **Response:**


```md-code__content
message SendMessageResponse {
    oneof payload {
      Task task = 1;
      Message msg = 2;
    }
}

```


- **URL:** `/v1/message:send`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
{
    message: Message,
    configuration?: MessageSendConfiguration,
    metadata?: { [key: string]: any }
}

```

- **Response:**


```md-code__content
// Returns one of a message or a task
{
    message?: Message,
    task?: Task
}

```


The `error` response for all transports in case of failure is a [`JSONRPCError`](https://a2a-protocol.org/latest/specification/#612-jsonrpcerror-object) or equivalent.

#### 7.1.1. `MessageSendParams` Object [¶](https://a2a-protocol.org/latest/specification/\#711-messagesendparams-object "Permanent link")

```md-code__content
/**
 * Defines the parameters for a request to send a message to an agent. This can be used
 * to create a new task, continue an existing one, or restart a task.
 */
export interface MessageSendParams {
  /** The message object being sent to the agent. */
  message: Message;
  /** Optional configuration for the send request. */
  configuration?: MessageSendConfiguration;
  /** Optional metadata for extensions. */
  metadata?: {
    [key: string]: any;
  };
}

/**
 * Defines configuration options for a `message/send` or `message/stream` request.
 */
export interface MessageSendConfiguration {
  /** A list of output MIME types the client is prepared to accept in the response. */
  acceptedOutputModes?: string[];
  /** The number of most recent messages from the task's history to retrieve in the response. */
  historyLength?: number;
  /** Configuration for the agent to send push notifications for updates after the initial response. */
  pushNotificationConfig?: PushNotificationConfig;
  /** If true, the client will wait for the task to complete. The server may reject this if the task is long-running. */
  blocking?: boolean;
}

```

### 7.2. `message/stream` [¶](https://a2a-protocol.org/latest/specification/\#72-messagestream "Permanent link")

Sends a message to an agent to initiate/continue a task AND subscribes the client to real-time updates for that task via Server-Sent Events (SSE). This method requires the server to have `AgentCard.capabilities.streaming: true`. Just like `message/send`, a task which has reached a terminal state (completed, canceled, rejected, or failed) can't be restarted. Sending a message to such a task will result in an error. For more information, refer to the [Life of a Task guide](https://a2a-protocol.org/latest/topics/life-of-a-task/).

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_1)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_1)[REST](https://a2a-protocol.org/latest/specification/#rest_1)

- **URL:** `message/stream`
- **HTTP Method:** `POST`
- **Payload**: [`MessageSendParams`](https://a2a-protocol.org/latest/specification/#711-messagesendparams-object) (same as `message/send`)
- **Response**: A stream of Server-Sent Events. Each SSE `data` field contains a [`SendStreamingMessageResponse`](https://a2a-protocol.org/latest/specification/#721-sendstreamingmessageresponse-object)

- **URL:** `SendStreamingMessage`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message SendMessageRequest {
    Message msg = 1;
    SendMessageConfiguration configuration = 2;
}

```

- **Response:**


```md-code__content
message StreamResponse {
    oneof payload {
      Task task;
      Message msg;
      TaskStatusUpdateEvent status_update;
      TaskArtifactUpdateEvent artifact_update;
    }
}

```


- **URL:** `/v1/message:stream`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
{
    message: Message,
    configuration?: MessageSendConfiguration,
    metadata?: { [key: string]: any }
}

```

- **Response:**


```md-code__content
{
    message?: Message
    task?: Task
    statusUpdate?: TaskStatusUpdateEvent
    artifactUpdate?: TaskArtifactUpdateEvent
}

```


#### 7.2.1. `SendStreamingMessageResponse` Object [¶](https://a2a-protocol.org/latest/specification/\#721-sendstreamingmessageresponse-object "Permanent link")

This is the structure of the JSON object found in the `data` field of each Server-Sent Event sent by the server for a `message/stream` request or `tasks/resubscribe` request.

```md-code__content
/**
 * Represents a JSON-RPC response for the `message/stream` method.
 */
export type SendStreamingMessageResponse =
  | SendStreamingMessageSuccessResponse
  | JSONRPCErrorResponse;

/**
 * Represents a successful JSON-RPC response for the `message/stream` method.
 * The server may send multiple response objects for a single request.
 */
export interface SendStreamingMessageSuccessResponse
  extends JSONRPCSuccessResponse {
  /** The result, which can be a Message, Task, or a streaming update event. */
  result: Message | Task | TaskStatusUpdateEvent | TaskArtifactUpdateEvent;
}

```

#### 7.2.2. `TaskStatusUpdateEvent` Object [¶](https://a2a-protocol.org/latest/specification/\#722-taskstatusupdateevent-object "Permanent link")

Carries information about a change in the task's status during streaming. This is one of the possible `result` types in a `SendStreamingMessageSuccessResponse`.

```md-code__content
/**
 * An event sent by the agent to notify the client of a change in a task's status.
 * This is typically used in streaming or subscription models.
 */
export interface TaskStatusUpdateEvent {
  /** The ID of the task that was updated. */
  taskId: string;
  /** The context ID associated with the task. */
  contextId: string;
  /** The type of this event, used as a discriminator. Always 'status-update'. */
  readonly kind: "status-update";
  /** The new status of the task. */
  status: TaskStatus;
  /** If true, this is the final event in the stream for this interaction. */
  final: boolean;
  /** Optional metadata for extensions. */
  metadata?: {
    [key: string]: any;
  };
}

```

#### 7.2.3. `TaskArtifactUpdateEvent` Object [¶](https://a2a-protocol.org/latest/specification/\#723-taskartifactupdateevent-object "Permanent link")

Carries a new or updated artifact (or a chunk of an artifact) generated by the task during streaming. This is one of the possible `result` types in a `SendTaskStreamingResponse`.

```md-code__content
/**
 * An event sent by the agent to notify the client that an artifact has been
 * generated or updated. This is typically used in streaming models.
 */
export interface TaskArtifactUpdateEvent {
  /** The ID of the task this artifact belongs to. */
  taskId: string;
  /** The context ID associated with the task. */
  contextId: string;
  /** The type of this event, used as a discriminator. Always 'artifact-update'. */
  readonly kind: "artifact-update";
  /** The artifact that was generated or updated. */
  artifact: Artifact;
  /** If true, the content of this artifact should be appended to a previously sent artifact with the same ID. */
  append?: boolean;
  /** If true, this is the final chunk of the artifact. */
  lastChunk?: boolean;
  /** Optional metadata for extensions. */
  metadata?: {
    [key: string]: any;
  };
}

```

### 7.3. `tasks/get` [¶](https://a2a-protocol.org/latest/specification/\#73-tasksget "Permanent link")

Retrieves the current state (including status, artifacts, and optionally history) of a previously initiated task. This is typically used for polling the status of a task initiated with `message/send`, or for fetching the final state of a task after being notified via a push notification or after an SSE stream has ended.

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_2)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_2)[REST](https://a2a-protocol.org/latest/specification/#rest_2)

- **URL:** `tasks/get`
- **HTTP Method:** `POST`
- **Payload**: [`TaskQueryParams`](https://a2a-protocol.org/latest/specification/#731-taskqueryparams-object)
- **Response**: `Task`

- **URL:** `GetTask`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message GetTaskRequest {
    // name=tasks/{id}
    string name;
    int32 history_length;
}

```

- **Response**: `Task`

- **URL:** `v1/tasks/{id}?historyLength={historyLength}`
- **HTTP Method:** `GET`
- **Payload:** None
- **Response**: `Task`

#### 7.3.1. `TaskQueryParams` Object [¶](https://a2a-protocol.org/latest/specification/\#731-taskqueryparams-object "Permanent link")

```md-code__content
/**
 * Defines parameters for querying a task, with an option to limit history length.
 */
export interface TaskQueryParams extends TaskIdParams {
  /** The number of most recent messages from the task's history to retrieve. */
  historyLength?: number;
}

```

### `tasks/list` [¶](https://a2a-protocol.org/latest/specification/\#taskslist "Permanent link")

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_3)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_3)[REST](https://a2a-protocol.org/latest/specification/#rest_3)

- N/A

- **URL:** `ListTask`
- **HTTP Method:** `GET`
- **Payload:**


```md-code__content
{}

```

- **Response**: `repeated Task`

- **URL:** `/v1/tasks`
- **HTTP Method:** `GET`
- **Payload:**


```md-code__content
{}

```

- **Response**: `[Task]`

### 7.4. `tasks/cancel` [¶](https://a2a-protocol.org/latest/specification/\#74-taskscancel "Permanent link")

Requests the cancellation of an ongoing task. The server will attempt to cancel the task, but success is not guaranteed (e.g., the task might have already completed or failed, or cancellation might not be supported at its current stage).

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_4)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_4)[REST](https://a2a-protocol.org/latest/specification/#rest_4)

- **URL:** `tasks/cancel`
- **HTTP Method:** `POST`
- **Payload**: [`TaskIdParams`](https://a2a-protocol.org/latest/specification/#741-taskidparams-object-for-taskscancel-and-taskspushnotificationconfigget)
- **Response**: `Task`

- **URL:** `CancelTask`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message CancelTaskRequest{
    // name=tasks/{id}
    string name;
}

```

- **Response**: `Task`

- **URL:** `/v1/tasks/{id}:cancel`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
{
    name: string
}

```

- **Response**: `Task`

#### 7.4.1. `TaskIdParams` Object (for `tasks/cancel` and `tasks/pushNotificationConfig/get`) [¶](https://a2a-protocol.org/latest/specification/\#741-taskidparams-object-for-taskscancel-and-taskspushnotificationconfigget "Permanent link")

A simple object containing just the task ID and optional metadata.

```md-code__content
/**
 * Defines parameters containing a task ID, used for simple task operations.
 */
export interface TaskIdParams {
  /** The unique identifier of the task. */
  id: string;
  /** Optional metadata associated with the request. */
  metadata?: {
    [key: string]: any;
  };
}

```

### 7.5. `tasks/pushNotificationConfig/set` [¶](https://a2a-protocol.org/latest/specification/\#75-taskspushnotificationconfigset "Permanent link")

Sets or updates the push notification configuration for a specified task. This allows the client to tell the server where and how to send asynchronous updates for the task. Requires the server to have `AgentCard.capabilities.pushNotifications: true`.

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_5)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_5)[REST](https://a2a-protocol.org/latest/specification/#rest_5)

- **URL:** `tasks/pushNotificationConfig/set`
- **HTTP Method:** `POST`
- **Payload**: [`TaskPushNotificationConfig`](https://a2a-protocol.org/latest/specification/#610-taskpushnotificationconfig-object)
- **Response**: [`TaskPushNotificationConfig`](https://a2a-protocol.org/latest/specification/#610-taskpushnotificationconfig-object)

- **URL:** `CreateTaskPushNotification`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message SetTaskPushNotificationRequest {
    TaskPushNotificationConfig config = 1;
}

```

- **Response**: `TaskPushNotificationConfig`

- **URL:** `/v1/tasks/{id}/pushNotificationConfigs`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
{
    config: TaskPushNotificationConfig
}

```

- **Response**: `TaskPushNotificationConfig`

### 7.6. `tasks/pushNotificationConfig/get` [¶](https://a2a-protocol.org/latest/specification/\#76-taskspushnotificationconfigget "Permanent link")

Retrieves the current push notification configuration for a specified task. Requires the server to have `AgentCard.capabilities.pushNotifications: true`.

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_6)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_6)[REST](https://a2a-protocol.org/latest/specification/#rest_6)

- **URL:** `tasks/pushNotificationConfig/get`
- **HTTP Method:** `POST`
- **Payload**: [`GetTaskPushNotificationConfigParams`](https://a2a-protocol.org/latest/specification/#761-gettaskpushnotificationconfigparams-object-taskspushnotificationconfigget)
- **Response**: [`TaskPushNotificationConfig`](https://a2a-protocol.org/latest/specification/#610-taskpushnotificationconfig-object)

- **URL:** `GetTaskPushNotification`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message TaskSubscriptionRequest {
    // name=tasks/{id}/pushNotification/{id}
    string name;
}

```

- **Response**: `TaskPushNotificationConfig`

- **URL:** `/v1/tasks/{taskId}/pushNotificationConfigs/{configId}`
- **HTTP Method:** `GET`
- **Payload:** None
- **Response**: `TaskPushNotificationConfig`

**Response `error` type (on failure)**: [`JSONRPCError`](https://a2a-protocol.org/latest/specification/#612-jsonrpcerror-object) (e.g., [`PushNotificationNotSupportedError`](https://a2a-protocol.org/latest/specification/#82-a2a-specific-errors), [`TaskNotFoundError`](https://a2a-protocol.org/latest/specification/#82-a2a-specific-errors)).

#### 7.6.1. `GetTaskPushNotificationConfigParams` Object ( `tasks/pushNotificationConfig/get`) [¶](https://a2a-protocol.org/latest/specification/\#761-gettaskpushnotificationconfigparams-object-taskspushnotificationconfigget "Permanent link")

A object for fetching the push notification configuration for a task.

```md-code__content
/**
 * Defines parameters for fetching a specific push notification configuration for a task.
 */
export interface GetTaskPushNotificationConfigParams extends TaskIdParams {
  /** The ID of the push notification configuration to retrieve. */
  pushNotificationConfigId?: string;
}

```

### 7.7. `tasks/pushNotificationConfig/list` [¶](https://a2a-protocol.org/latest/specification/\#77-taskspushnotificationconfiglist "Permanent link")

Retrieves the associated push notification configurations for a specified task. Requires the server to have `AgentCard.capabilities.pushNotifications: true`.

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_7)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_7)[REST](https://a2a-protocol.org/latest/specification/#rest_7)

- **URL:** `tasks/pushNotificationConfig/list`
- **HTTP Method:** `POST`
- **Payload:** [`ListTaskPushNotificationConfigParams`](https://a2a-protocol.org/latest/specification/#771-listtaskpushnotificationconfigparams-object-taskspushnotificationconfiglist)
- **Response**: `TaskPushNotificationConfig[]`

- **URL:** `ListTaskPushNotification`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message ListTaskPushNotificationRequest {
    // parent=tasks/{id}
    string parent = 1;
}

```

- **Response**: `repeated TaskPushNotificationConfig`

- **URL:** `/v1/tasks/{id}/pushNotificationConfigs`
- **HTTP Method:** `GET`
- **Payload:**: None
- **Response**: `[TaskPushNotificationConfig]`

#### 7.7.1. `ListTaskPushNotificationConfigParams` Object ( `tasks/pushNotificationConfig/list`) [¶](https://a2a-protocol.org/latest/specification/\#771-listtaskpushnotificationconfigparams-object-taskspushnotificationconfiglist "Permanent link")

A object for fetching the push notification configurations for a task.

```md-code__content
/**
 * Defines parameters for listing all push notification configurations associated with a task.
 */
export interface ListTaskPushNotificationConfigParams extends TaskIdParams {}

```

### 7.8. `tasks/pushNotificationConfig/delete` [¶](https://a2a-protocol.org/latest/specification/\#78-taskspushnotificationconfigdelete "Permanent link")

Deletes an associated push notification configuration for a task. Requires the server to have `AgentCard.capabilities.pushNotifications: true`.

- **Request `params` type**: [`DeleteTaskPushNotificationConfigParams`](https://a2a-protocol.org/latest/specification/#781-deletetaskpushnotificationconfigparams-object-taskspushnotificationconfigdelete)
- **Response `result` type (on success)**: \[ `null`\]
- **Response `error` type (on failure)**: [`JSONRPCError`](https://a2a-protocol.org/latest/specification/#612-jsonrpcerror-object) (e.g., [`PushNotificationNotSupportedError`](https://a2a-protocol.org/latest/specification/#82-a2a-specific-errors), [`TaskNotFoundError`](https://a2a-protocol.org/latest/specification/#82-a2a-specific-errors)).

#### 7.8.1. `DeleteTaskPushNotificationConfigParams` Object ( `tasks/pushNotificationConfig/delete`) [¶](https://a2a-protocol.org/latest/specification/\#781-deletetaskpushnotificationconfigparams-object-taskspushnotificationconfigdelete "Permanent link")

A object for deleting an associated push notification configuration for a task.

```md-code__content
/**
 * Defines parameters for deleting a specific push notification configuration for a task.
 */
export interface DeleteTaskPushNotificationConfigParams extends TaskIdParams {
  /** The ID of the push notification configuration to delete. */
  pushNotificationConfigId: string;
}

```

### 7.9. `tasks/resubscribe` [¶](https://a2a-protocol.org/latest/specification/\#79-tasksresubscribe "Permanent link")

Allows a client to reconnect to an SSE stream for an ongoing task after a previous connection (from `message/stream` or an earlier `tasks/resubscribe`) was interrupted. Requires the server to have `AgentCard.capabilities.streaming: true`.

The purpose is to resume receiving _subsequent_ updates. The server's behavior regarding events missed during the disconnection period (e.g., whether it attempts to backfill some missed events or only sends new ones from the point of resubscription) is implementation-dependent and not strictly defined by this specification.

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_8)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_8)[REST](https://a2a-protocol.org/latest/specification/#rest_8)

- **URL:** `tasks/resubscribe`
- **HTTP Method:** `POST`
- **Payload**: [`TaskIdParams`](https://a2a-protocol.org/latest/specification/#741-taskidparams-object-for-taskscancel-and-taskspushnotificationconfigget)
- **Response**: A stream of Server-Sent Events. Each SSE `data` field contains a [`SendStreamingMessageResponse`](https://a2a-protocol.org/latest/specification/#721-sendstreamingmessageresponse-object)

- **URL:** `TaskSubscription`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
message TaskSubscriptionRequest{
    // name=tasks/{id}
    string name;
}

```

- **Response:**


```md-code__content
message StreamResponse {
    oneof payload {
      Task task;
      Message msg;
      TaskStatusUpdateEvent status_update;
      TaskArtifactUpdateEvent artifact_update;
    }
}

```


- **URL:** `/v1/tasks/{id}:subscribe`
- **HTTP Method:** `POST`
- **Payload:**


```md-code__content
{
    name: string
}

```

- **Response:**


```md-code__content
{
    message?: Message
    task?: Task
    statusUpdate?: TaskStatusUpdateEvent
    artifactUpdate?: TaskArtifactUpdateEvent
}

```


### 7.10. `agent/getAuthenticatedExtendedCard` [¶](https://a2a-protocol.org/latest/specification/\#710-agentgetauthenticatedextendedcard "Permanent link")

Retrieves a potentially more detailed version of the Agent Card after the client has authenticated. This endpoint is available only if `AgentCard.supportsAuthenticatedExtendedCard` is `true`.

- **Authentication**: The client **MUST** authenticate the request using one of the schemes declared in the public `AgentCard.securitySchemes` and `AgentCard.security` fields.
- **Response `result` type (on success)**: `AgentCard` (A complete Agent Card object, which may contain additional details or skills not present in the public card).
- **Response `error` type (on failure)**: Standard HTTP error codes.
  - `401 Unauthorized`: Authentication failed (missing or invalid credentials). The server **SHOULD** include a `WWW-Authenticate` header.

[JSON-RPC](https://a2a-protocol.org/latest/specification/#json-rpc_9)[gRPC](https://a2a-protocol.org/latest/specification/#grpc_9)[REST](https://a2a-protocol.org/latest/specification/#rest_9)

- **URL:** `agent/getAuthenticatedExtendedCard`
- **HTTP Method:** `POST`
- **Payload:** None
- **Response:** `AgentCard`

- **URL:** `GetAgentCard`
- **HTTP Method:** `POST`
- **Payload:** None
- **Response:** `AgentCard`

- **URL:** `/v1/card`
- **HTTP Method:** `GET`
- **Payload:** None
- **Response:** `AgentCard`

Clients retrieving this authenticated card **SHOULD** replace their cached public Agent Card with the content received from this endpoint for the duration of their authenticated session or until the card's version changes.

```md-code__content
/**
 * Represents a successful JSON-RPC response for the `agent/getAuthenticatedExtendedCard` method.
 */
export interface GetAuthenticatedExtendedCardSuccessResponse
  extends JSONRPCSuccessResponse {
  /** The result is an Agent Card object. */
  result: AgentCard;
}

```

## 8\. Error Handling [¶](https://a2a-protocol.org/latest/specification/\#8-error-handling "Permanent link")

A2A uses standard [JSON-RPC 2.0 error codes and structure](https://www.jsonrpc.org/specification#error_object) for reporting errors. Errors are returned in the `error` member of the `JSONRPCErrorResponse` object. See [`JSONRPCError` Object definition](https://a2a-protocol.org/latest/specification/#612-jsonrpcerror-object).

### 8.1. Standard JSON-RPC Errors [¶](https://a2a-protocol.org/latest/specification/\#81-standard-json-rpc-errors "Permanent link")

These are standard codes defined by the JSON-RPC 2.0 specification.

| Code | JSON-RPC Spec Meaning | Typical A2A `message` | Description |
| --- | --- | --- | --- |
| `-32700` | Parse error | Invalid JSON payload | Server received JSON that was not well-formed. |
| `-32600` | Invalid Request | Invalid JSON-RPC Request | The JSON payload was valid JSON, but not a valid JSON-RPC Request object. |
| `-32601` | Method not found | Method not found | The requested A2A RPC `method` (e.g., `"tasks/foo"`) does not exist or is not supported. |
| `-32602` | Invalid params | Invalid method parameters | The `params` provided for the method are invalid (e.g., wrong type, missing required field). |
| `-32603` | Internal error | Internal server error | An unexpected error occurred on the server during processing. |
| `-32000` to `-32099` | Server error | _(Server-defined)_ | Reserved for implementation-defined server-errors. A2A-specific errors use this range. |

### 8.2. A2A-Specific Errors [¶](https://a2a-protocol.org/latest/specification/\#82-a2a-specific-errors "Permanent link")

These are custom error codes defined within the JSON-RPC server error range ( `-32000` to `-32099`) to provide more specific feedback about A2A-related issues. Servers **SHOULD** use these codes where applicable.

| Code | Error Name (Conceptual) | Typical `message` string | Description |
| --- | --- | --- | --- |
| `-32001` | `TaskNotFoundError` | Task not found | The specified task `id` does not correspond to an existing or active task. It might be invalid, expired, or already completed and purged. |
| `-32002` | `TaskNotCancelableError` | Task cannot be canceled | An attempt was made to cancel a task that is not in a cancelable state (e.g., it has already reached a terminal state like `completed`, `failed`, or `canceled`). |
| `-32003` | `PushNotificationNotSupportedError` | Push Notification is not supported | Client attempted to use push notification features (e.g., `tasks/pushNotificationConfig/set`) but the server agent does not support them (i.e., `AgentCard.capabilities.pushNotifications` is `false`). |
| `-32004` | `UnsupportedOperationError` | This operation is not supported | The requested operation or a specific aspect of it (perhaps implied by parameters) is not supported by this server agent implementation. Broader than just method not found. |
| `-32005` | `ContentTypeNotSupportedError` | Incompatible content types | A [Media Type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types) provided in the request's `message.parts` (or implied for an artifact) is not supported by the agent or the specific skill being invoked. |
| `-32006` | `InvalidAgentResponseError` | Invalid agent response type | Agent generated an invalid response for the requested method |
| `-32007` | `AuthenticatedExtendedCardNotConfiguredError` | Authenticated Extended Card not configured | The agent does not have an Authenticated Extended Card configured. |

Servers MAY define additional error codes within the `-32000` to `-32099` range for more specific scenarios not covered above, but they **SHOULD** document these clearly. The `data` field of the `JSONRPCError` object can be used to provide more structured details for any error.

## 9\. Common Workflows & Examples [¶](https://a2a-protocol.org/latest/specification/\#9-common-workflows-examples "Permanent link")

This section provides illustrative JSON examples of common A2A interactions. Timestamps, context IDs, and request/response IDs are for demonstration purposes. For brevity, some optional fields might be omitted if not central to the example.

### 9.1. Fetching Authenticated Extended Agent Card [¶](https://a2a-protocol.org/latest/specification/\#91-fetching-authenticated-extended-agent-card "Permanent link")

**Scenario:** A client discovers a public Agent Card indicating support for an authenticated extended card and wants to retrieve the full details.

1. **Client fetches the public Agent Card:**

```md-code__content
GET https://example.com/.well-known/agent-card.json

```

_Server responds with the public Agent Card (like the example in Section 5.6), including `supportsAuthenticatedExtendedCard: true` (at the root level) and `securitySchemes`._

1. **Client identifies required authentication from the public card.**

2. **Client obtains necessary credentials out-of-band (e.g., performs OAuth 2.0 flow with Google, resulting in an access token).**

3. **Client fetches the authenticated extended Agent Card using `agent/getAuthenticatedExtendedCard` request:**


```md-code__content
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "agent/getAuthenticatedExtendedCard"
}

```

1. **Server authenticates and authorizes the request.**

2. **Server responds with the full Agent Card as the JSON-RPC result:**


### 9.2. Basic Execution (Synchronous / Polling Style) [¶](https://a2a-protocol.org/latest/specification/\#92-basic-execution-synchronous-polling-style "Permanent link")

**Scenario:** Client asks a simple question, and the agent responds quickly with a task

1. **Client sends a message using `message/send`:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "tell me a joke"\
        }\
      ],
      "messageId": "9229e770-767c-417b-a0b0-f0741243c589"
    },
    "metadata": {}
  }
}

```

1. **Server processes the request, creates a task and responds (task completes quickly)**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "id": "363422be-b0f9-4692-a24d-278670e7c7f1",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "status": {
      "state": "completed"
    },
    "artifacts": [\
      {\
        "artifactId": "9b6934dd-37e3-4eb1-8766-962efaab63a1",\
        "name": "joke",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "Why did the chicken cross the road? To get to the other side!"\
          }\
        ]\
      }\
    ],
    "history": [\
      {\
        "role": "user",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "tell me a joke"\
          }\
        ],\
        "messageId": "9229e770-767c-417b-a0b0-f0741243c589",\
        "taskId": "363422be-b0f9-4692-a24d-278670e7c7f1",\
        "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4"\
      }\
    ],
    "kind": "task",
    "metadata": {}
  }
}

```

_If the task were longer-running, the server might initially respond with `status.state: "working"`. The client would then periodically call `tasks/get` with params: `{"id": "363422be-b0f9-4692-a24d-278670e7c7f1"}` until the task reaches a terminal state._

**Scenario:** Client asks a simple question, and the agent responds quickly without a task

1. **Client sends a message using `message/send`:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "tell me a joke"\
        }\
      ],
      "messageId": "9229e770-767c-417b-a0b0-f0741243c589"
    },
    "metadata": {}
  }
}

```

1. **Server processes the request, responds quickly without a task**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "messageId": "363422be-b0f9-4692-a24d-278670e7c7f1",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "parts": [\
      {\
        "kind": "text",\
        "text": "Why did the chicken cross the road? To get to the other side!"\
      }\
    ],
    "kind": "message",
    "metadata": {}
  }
}

```

### 9.3. Streaming Task Execution (SSE) [¶](https://a2a-protocol.org/latest/specification/\#93-streaming-task-execution-sse "Permanent link")

**Scenario:** Client asks the agent to write a long paper describing an attached picture.

1. **Client sends a message and subscribes using `message/stream`:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "message/stream",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "write a long paper describing the attached pictures"\
        },\
        {\
          "kind": "file",\
          "file": {\
            "mimeType": "image/png",\
            "data": "<base64-encoded-content>"\
          }\
        }\
      ],
      "messageId": "bbb7dee1-cf5c-4683-8a6f-4114529da5eb"
    },
    "metadata": {}
  }
}

```

1. **Server responds with HTTP 200 OK, `Content-Type: text/event-stream`, and starts sending SSE events:**

_Event 1: Task status update - working_

```md-code__content
data: {
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "id": "225d6247-06ba-4cda-a08b-33ae35c8dcfa",
    "contextId": "05217e44-7e9f-473e-ab4f-2c2dde50a2b1",
    "status": {
      "state": "submitted",
      "timestamp":"2025-04-02T16:59:25.331844"
    },
    "history": [\
      {\
        "role": "user",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "write a long paper describing the attached pictures"\
          },\
          {\
            "kind": "file",\
            "file": {\
              "mimeType": "image/png",\
              "data": "<base64-encoded-content>"\
            }\
          }\
        ],\
        "messageId": "bbb7dee1-cf5c-4683-8a6f-4114529da5eb",\
        "taskId": "225d6247-06ba-4cda-a08b-33ae35c8dcfa",\
        "contextId": "05217e44-7e9f-473e-ab4f-2c2dde50a2b1"\
      }\
    ],
    "kind": "task",
    "metadata": {}
  }
}

data: {
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "taskId": "225d6247-06ba-4cda-a08b-33ae35c8dcfa",
    "contextId": "05217e44-7e9f-473e-ab4f-2c2dde50a2b1",
    "artifact": {
      "artifactId": "9b6934dd-37e3-4eb1-8766-962efaab63a1",
      "parts": [\
        {"kind":"text", "text": "<section 1...>"}\
      ]
    },
    "append": false,
    "lastChunk": false,
    "kind":"artifact-update"
  }
}

data: {
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "taskId": "225d6247-06ba-4cda-a08b-33ae35c8dcfa",
    "contextId": "05217e44-7e9f-473e-ab4f-2c2dde50a2b1",
    "artifact": {
      "artifactId": "9b6934dd-37e3-4eb1-8766-962efaab63a1",
      "parts": [\
        {"kind":"text", "text": "<section 2...>"}\
      ],
    },
    "append": true,
    "lastChunk": false,
    "kind":"artifact-update"
  }
}

data: {
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "taskId": "225d6247-06ba-4cda-a08b-33ae35c8dcfa",
    "contextId": "05217e44-7e9f-473e-ab4f-2c2dde50a2b1",
    "artifact": {
      "artifactId": "9b6934dd-37e3-4eb1-8766-962efaab63a1",
      "parts": [\
        {"kind":"text", "text": "<section 3...>"}\
      ]
    },
    "append": true,
    "lastChunk": true,
    "kind":"artifact-update"
  }
}

data: {
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "taskId": "225d6247-06ba-4cda-a08b-33ae35c8dcfa",
    "contextId": "05217e44-7e9f-473e-ab4f-2c2dde50a2b1",
    "status": {
      "state": "completed",
      "timestamp":"2025-04-02T16:59:35.331844"
    },
    "final": true,
    "kind":"status-update"
  }
}

```

_(Server closes the SSE connection after the `final:true` event)._

### 9.4. Multi-Turn Interaction (Input Required) [¶](https://a2a-protocol.org/latest/specification/\#94-multi-turn-interaction-input-required "Permanent link")

**Scenario:** Client wants to book a flight, and the agent needs more information.

1. **Client sends a message using `message/send`:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [{ "kind": "text", "text": "I'd like to book a flight." }]
    },
    "messageId": "c53ba666-3f97-433c-a87b-6084276babe2"
  }
}

```

1. **Server responds, task state is `input-required`:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "result": {
    "id": "3f36680c-7f37-4a5f-945e-d78981fafd36",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "status": {
      "state": "input-required",
      "message": {
        "role": "agent",
        "parts": [\
          {\
            "kind": "text",\
            "text": "Sure, I can help with that! Where would you like to fly to, and from where? Also, what are your preferred travel dates?"\
          }\
        ],
        "messageId": "c2e1b2dd-f200-4b04-bc22-1b0c65a1aad2",
        "taskId": "3f36680c-7f37-4a5f-945e-d78981fafd36",
        "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4"
      },
      "timestamp": "2024-03-15T10:10:00Z"
    },
    "history": [\
      {\
        "role": "user",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "I'd like to book a flight."\
          }\
        ],\
        "messageId": "c53ba666-3f97-433c-a87b-6084276babe2",\
        "taskId": "3f36680c-7f37-4a5f-945e-d78981fafd36",\
        "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4"\
      }\
    ],
    "kind": "task"
  }
}

```

1. **Client `message/send` (providing the requested input, using the _same_ task ID):**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-004",
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "I want to fly from New York (JFK) to London (LHR) around October 10th, returning October 17th."\
        }\
      ],
      "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
      "taskId": "3f36680c-7f37-4a5f-945e-d78981fafd36",
      "messageId": "0db1d6c4-3976-40ed-b9b8-0043ea7a03d3"
    },
    "configuration": {
      "blocking": true
    }
  }
}

```

1. **Server processes the new input and responds (e.g., task completed or more input needed):**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-004",
  "result": {
    "id": "3f36680c-7f37-4a5f-945e-d78981fafd36",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "status": {
      "state": "completed",
      "message": {
        "role": "agent",
        "parts": [\
          {\
            "kind": "text",\
            "text": "Okay, I've found a flight for you. Confirmation XYZ123. Details are in the artifact."\
          }\
        ]
      }
    },
    "artifacts": [\
      {\
        "artifactId": "9b6934dd-37e3-4eb1-8766-962efaab63a1",\
        "name": "FlightItinerary.json",\
        "parts": [\
          {\
            "kind": "data",\
            "data": {\
              "confirmationId": "XYZ123",\
              "from": "JFK",\
              "to": "LHR",\
              "departure": "2024-10-10T18:00:00Z",\
              "arrival": "2024-10-11T06:00:00Z",\
              "returnDeparture": "..."\
            }\
          }\
        ]\
      }\
    ],
    "history": [\
      {\
        "role": "user",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "I'd like to book a flight."\
          }\
        ],\
        "messageId": "c53ba666-3f97-433c-a87b-6084276babe2",\
        "taskId": "3f36680c-7f37-4a5f-945e-d78981fafd36",\
        "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4"\
      },\
      {\
        "role": "agent",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "Sure, I can help with that! Where would you like to fly to, and from where? Also, what are your preferred travel dates?"\
          }\
        ],\
        "messageId": "c2e1b2dd-f200-4b04-bc22-1b0c65a1aad2",\
        "taskId": "3f36680c-7f37-4a5f-945e-d78981fafd36",\
        "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4"\
      },\
      {\
        "role": "user",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "I want to fly from New York (JFK) to London (LHR) around October 10th, returning October 17th."\
          }\
        ],\
        "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",\
        "taskId": "3f36680c-7f37-4a5f-945e-d78981fafd36",\
        "messageId": "0db1d6c4-3976-40ed-b9b8-0043ea7a03d3"\
      }\
    ],
    "kind": "task",
    "metadata": {}
  }
}

```

### 9.5. Push Notification Setup and Usage [¶](https://a2a-protocol.org/latest/specification/\#95-push-notification-setup-and-usage "Permanent link")

**Scenario:** Client requests a long-running report generation and wants to be notified via webhook when it's done.

1. **Client `message/send` with `pushNotification` config:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-005",
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "Generate the Q1 sales report. This usually takes a while. Notify me when it's ready."\
        }\
      ],
      "messageId": "6dbc13b5-bd57-4c2b-b503-24e381b6c8d6"
    },
    "configuration": {
      "pushNotificationConfig": {
        "url": "https://client.example.com/webhook/a2a-notifications",
        "token": "secure-client-token-for-task-aaa",
        "authentication": {
          "schemes": ["Bearer"]
          // Assuming server knows how to get a Bearer token for this webhook audience,
          // or this implies the webhook is public/uses the 'token' for auth.
          // 'credentials' could provide more specifics if needed by the server.
        }
      }
    }
  }
}

```

1. **Server acknowledges the task (e.g., status `submitted` or `working`):**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-005",
  "result": {
    "id": "43667960-d455-4453-b0cf-1bae4955270d",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "status": { "state": "submitted", "timestamp": "2024-03-15T11:00:00Z" }
    // ... other fields ...
  }
}

```

1. **(Later) A2A Server completes the task and POSTs a notification to `https://client.example.com/webhook/a2a-notifications`:**

2. **HTTP Headers might include:**
   - `Authorization: Bearer <server_jwt_for_webhook_audience>` (if server authenticates to webhook)
   - `Content-Type: application/json`
   - `X-A2A-Notification-Token: secure-client-token-for-task-aaa`
3. **HTTP Body (Task object is sent as JSON payload):**

```md-code__content
{
  "id": "43667960-d455-4453-b0cf-1bae4955270d",
  "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
  "status": { "state": "completed", "timestamp": "2024-03-15T18:30:00Z" },
  "kind": "task"
  // ... other fields ...
}

```

1. **Client's Webhook Service:**

2. Receives the POST.

3. Validates the `Authorization` header (if applicable).
4. Validates the `X-A2A-Notification-Token`.
5. Internally processes the notification (e.g., updates application state, notifies end user).

### 9.6. File Exchange (Upload and Download) [¶](https://a2a-protocol.org/latest/specification/\#96-file-exchange-upload-and-download "Permanent link")

**Scenario:** Client sends an image for analysis, and the agent returns a modified image.

1. **Client `message/send` with a `FilePart` (uploading image bytes):**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-007",
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "Analyze this image and highlight any faces."\
        },\
        {\
          "kind": "file",\
          "file": {\
            "name": "input_image.png",\
            "mimeType": "image/png",\
            "bytes": "iVBORw0KGgoAAAANSUhEUgAAAAUA..." // Base64 encoded image data\
          }\
        }\
      ],
      "messageId": "6dbc13b5-bd57-4c2b-b503-24e381b6c8d6"
    }
  }
}

```

1. **Server processes the image and responds with a `FilePart` in an artifact (e.g., providing a URI to the modified image):**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": "req-007",
  "result": {
    "id": "43667960-d455-4453-b0cf-1bae4955270d",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "status": { "state": "completed", "timestamp": "2024-03-15T12:05:00Z" },
    "artifacts": [\
      {\
        "artifactId": "9b6934dd-37e3-4eb1-8766-962efaab63a1",\
        "name": "processed_image_with_faces.png",\
        "parts": [\
          {\
            "kind": "file",\
            "file": {\
              "name": "output.png",\
              "mimeType": "image/png",\
              // Server might provide a URI to a temporary storage location\
              "uri": "https://storage.example.com/processed/task-bbb/output.png?token=xyz"\
              // Or, alternatively, it could return bytes directly:\
              // "bytes": "ASEDGhw0KGgoAAAANSUhEUgAA..."\
            }\
          }\
        ]\
      }\
    ],
    "kind": "task"
  }
}

```

### 9.7. Structured Data Exchange (Requesting and Providing JSON) [¶](https://a2a-protocol.org/latest/specification/\#97-structured-data-exchange-requesting-and-providing-json "Permanent link")

**Scenario:** Client asks for a list of open support tickets in a specific JSON format.

1. **Client `message/send`, `Part.metadata` hints at desired output schema/Media Type:** _(Note: A2A doesn't formally standardize schema negotiation in v0.2.0, but `metadata` can be used for such hints by convention between client/server)._

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 9,
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [\
        {\
          "kind": "text",\
          "text": "Show me a list of my open IT tickets",\
          "metadata": {\
            "mimeType": "application/json",\
            "schema": {\
              "type": "array",\
              "items": {\
                "type": "object",\
                "properties": {\
                  "ticketNumber": { "type": "string" },\
                  "description": { "type": "string" }\
                }\
              }\
            }\
          }\
        }\
      ],
      "messageId": "85b26db5-ffbb-4278-a5da-a7b09dea1b47"
    },
    "metadata": {}
  }
}

```

1. **Server responds with structured JSON data:**

```md-code__content
{
  "jsonrpc": "2.0",
  "id": 9,
  "result": {
    "id": "d8c6243f-5f7a-4f6f-821d-957ce51e856c",
    "contextId": "c295ea44-7543-4f78-b524-7a38915ad6e4",
    "status": {
      "state": "completed",
      "timestamp": "2025-04-17T17:47:09.680794"
    },
    "artifacts": [\
      {\
        "artifactId": "c5e0382f-b57f-4da7-87d8-b85171fad17c",\
        "parts": [\
          {\
            "kind": "text",\
            "text": "[{\"ticketNumber\":\"REQ12312\",\"description\":\"request for VPN access\"},{\"ticketNumber\":\"REQ23422\",\"description\":\"Add to DL - team-gcp-onboarding\"}]"\
          }\
        ]\
      }\
    ],
    "kind": "task"
  }
}

```

These examples illustrate the flexibility of A2A in handling various interaction patterns and data types. Implementers should refer to the detailed object definitions for all fields and constraints.

## 10\. Appendices [¶](https://a2a-protocol.org/latest/specification/\#10-appendices "Permanent link")

### 10.1. Relationship to MCP (Model Context Protocol) [¶](https://a2a-protocol.org/latest/specification/\#101-relationship-to-mcp-model-context-protocol "Permanent link")

A2A and MCP are complementary protocols designed for different aspects of agentic systems:

- **[Model Context Protocol (MCP)](https://modelcontextprotocol.io/):** Focuses on standardizing how AI models and agents connect to and interact with **tools, APIs, data sources, and other external resources.** It defines structured ways to describe tool capabilities (like function calling in LLMs), pass inputs, and receive structured outputs. Think of MCP as the "how-to" for an agent to _use_ a specific capability or access a resource.
- **Agent2Agent Protocol (A2A):** Focuses on standardizing how independent, often opaque, **AI agents communicate and collaborate with each other as peers.** A2A provides an application-level protocol for agents to discover each other, negotiate interaction modalities, manage shared tasks, and exchange conversational context or complex results. It's about how agents _partner_ or _delegate_ work.

**How they work together:**
An A2A Client agent might request an A2A Server agent to perform a complex task. The Server agent, in turn, might use MCP to interact with several underlying tools, APIs, or data sources to gather information or perform actions necessary to fulfill the A2A task.

For a more detailed comparison, see the [A2A and MCP guide](https://a2a-protocol.org/latest/topics/a2a-and-mcp/).

### 10.2. Security Considerations Summary [¶](https://a2a-protocol.org/latest/specification/\#102-security-considerations-summary "Permanent link")

Security is a paramount concern in A2A. Key considerations include:

- **Transport Security:** Always use HTTPS with strong TLS configurations in production environments.
- **Authentication:**
  - Handled via standard HTTP mechanisms (e.g., `Authorization` header with Bearer tokens, API keys).
  - Requirements are declared in the `AgentCard`.
  - Credentials MUST be obtained out-of-band by the client.
  - A2A Servers MUST authenticate every request.
- **Authorization:**
  - A server-side responsibility based on the authenticated identity.
  - Implement the principle of least privilege.
  - Can be granular, based on skills, actions, or data.
- **Push Notification Security:**
  - Webhook URL validation (by the A2A Server sending notifications) is crucial to prevent SSRF.
  - Authentication of the A2A Server to the client's webhook is essential.
  - Authentication of the notification by the client's webhook receiver (verifying it came from the legitimate A2A Server and is relevant) is critical.
  - See the [Streaming & Asynchronous Operations guide](https://a2a-protocol.org/latest/topics/streaming-and-async/#security-considerations-for-push-notifications) for detailed push notification security.
- **Input Validation:** Servers MUST rigorously validate all RPC parameters and the content/structure of data in `Message` and `Artifact` parts to prevent injection attacks or processing errors.
- **Resource Management:** Implement rate limiting, concurrency controls, and resource limits to protect agents from abuse or overload.
- **Data Privacy:** Adhere to all applicable privacy regulations for data exchanged in `Message` and `Artifact` parts. Minimize sensitive data transfer.

For a comprehensive discussion, refer to the [Enterprise-Ready Features guide](https://a2a-protocol.org/latest/topics/enterprise-ready/).

## 11\. A2A Compliance Requirements [¶](https://a2a-protocol.org/latest/specification/\#11-a2a-compliance-requirements "Permanent link")

This section defines the normative requirements for A2A-compliant implementations.

### 11.1. Agent Compliance [¶](https://a2a-protocol.org/latest/specification/\#111-agent-compliance "Permanent link")

For an agent to be considered **A2A-compliant**, it **MUST**:

#### 11.1.1. Transport Support Requirements [¶](https://a2a-protocol.org/latest/specification/\#1111-transport-support-requirements "Permanent link")

- **Support at least one transport**: Agents **MUST** implement at least one transport protocols as defined in [Section 3.2](https://a2a-protocol.org/latest/specification/#32-supported-transport-protocols).
- **Expose Agent Card**: **MUST** provide a valid `AgentCard` document as defined in [Section 5](https://a2a-protocol.org/latest/specification/#5-agent-discovery-the-agent-card).
- **Declare transport capabilities**: **MUST** accurately declare all supported transports in the `AgentCard` using `preferredTransport` and `additionalInterfaces` fields following the requirements in [Section 5.6](https://a2a-protocol.org/latest/specification/#56-transport-declaration-and-url-relationships).

#### 11.1.2. Core Method Implementation [¶](https://a2a-protocol.org/latest/specification/\#1112-core-method-implementation "Permanent link")

**MUST** implement all of the following core methods via at least one supported transport:

- `message/send` \- Send messages and initiate tasks
- `tasks/get` \- Retrieve task status and results
- `tasks/cancel` \- Request task cancellation

#### 11.1.3. Optional Method Implementation [¶](https://a2a-protocol.org/latest/specification/\#1113-optional-method-implementation "Permanent link")

**MAY** implement the following optional methods:

- `message/stream` \- Streaming message interaction (requires `capabilities.streaming: true`)
- `tasks/resubscribe` \- Resume streaming for existing tasks (requires `capabilities.streaming: true`)
- `tasks/pushNotificationConfig/set` \- Configure push notifications (requires `capabilities.pushNotifications: true`)
- `tasks/pushNotificationConfig/get` \- Retrieve push notification config (requires `capabilities.pushNotifications: true`)
- `tasks/pushNotificationConfig/list` \- List push notification configs (requires `capabilities.pushNotifications: true`)
- `tasks/pushNotificationConfig/delete` \- Delete push notification config (requires `capabilities.pushNotifications: true`)
- `agent/authenticatedExtendedCard` \- Retrieve authenticated agent card (requires `supportsAuthenticatedExtendedCard: true`)

#### 11.1.4. Multi-Transport Compliance [¶](https://a2a-protocol.org/latest/specification/\#1114-multi-transport-compliance "Permanent link")

If an agent supports additional transports (gRPC, HTTP+JSON), it **MUST**:

- **Functional equivalence**: Provide identical functionality across all supported transports.
- **Consistent behavior**: Return semantically equivalent results for the same operations.
- **Transport-specific requirements**: Conform to all requirements defined in [Section 3.2](https://a2a-protocol.org/latest/specification/#32-supported-transport-protocols) for each supported transport.
- **Method mapping compliance**: Use the standard method mappings defined in [Section 3.5](https://a2a-protocol.org/latest/specification/#35-method-mapping-and-naming-conventions) for all supported transports.

#### 11.1.5. Data Format Compliance [¶](https://a2a-protocol.org/latest/specification/\#1115-data-format-compliance "Permanent link")

- **JSON-RPC structure**: **MUST** use valid JSON-RPC 2.0 request/response objects as defined in [Section 6.11](https://a2a-protocol.org/latest/specification/#611-json-rpc-structures).
- **A2A data objects**: **MUST** use the data structures defined in [Section 6](https://a2a-protocol.org/latest/specification/#6-protocol-data-objects) for all protocol entities.
- **Error handling**: **MUST** use the error codes defined in [Section 8](https://a2a-protocol.org/latest/specification/#8-error-handling).

### 11.2. Client Compliance [¶](https://a2a-protocol.org/latest/specification/\#112-client-compliance "Permanent link")

For a client to be considered **A2A-compliant**, it **MUST**:

#### 11.2.1. Transport Support [¶](https://a2a-protocol.org/latest/specification/\#1121-transport-support "Permanent link")

- **Multi-transport capability**: **MUST** be able to communicate with agents using at least one transport protocols.
- **Agent Card processing**: **MUST** be able to parse and interpret `AgentCard` documents.
- **Transport selection**: **MUST** be able to select an appropriate transport from the agent's declared capabilities following the rules defined in [Section 5.6.3](https://a2a-protocol.org/latest/specification/#563-client-transport-selection-rules).

#### 11.2.2. Protocol Implementation [¶](https://a2a-protocol.org/latest/specification/\#1122-protocol-implementation "Permanent link")

- **Core method usage**: **MUST** properly construct requests for at least `message/send` and `tasks/get` methods.
- **Error handling**: **MUST** properly handle all A2A error codes defined in [Section 8.2](https://a2a-protocol.org/latest/specification/#82-a2a-specific-errors).
- **Authentication**: **MUST** support at least one authentication method when interacting with agents that require authentication.

#### 11.2.3. Optional Client Features [¶](https://a2a-protocol.org/latest/specification/\#1123-optional-client-features "Permanent link")

Clients **MAY** implement:

- **Multi-transport support**: Support for gRPC and/or HTTP+JSON transports.
- **Streaming support**: Handle streaming methods and Server-Sent Events.
- **Push notification handling**: Serve as webhook endpoints for push notifications.
- **Extended Agent Cards**: Retrieve and use authenticated extended agent cards.

### 11.3. Compliance Testing [¶](https://a2a-protocol.org/latest/specification/\#113-compliance-testing "Permanent link")

Implementations **SHOULD** validate compliance through:

- **Transport interoperability**: Test communication with agents using different transport implementations.
- **Method mapping verification**: Verify that all supported transports use the correct method names and URL patterns as defined in [Section 3.5](https://a2a-protocol.org/latest/specification/#35-method-mapping-and-naming-conventions).
- **Error handling**: Verify proper handling of all defined error conditions.
- **Data format validation**: Ensure JSON schemas match the TypeScript type definitions in [`types/src/types.ts`](https://a2a-protocol.org/latest/specification/types/src/types.ts).
- **Multi-transport consistency**: For multi-transport agents, verify functional equivalence across all supported transports.
