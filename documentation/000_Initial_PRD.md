Product Requirements Document: ComplyArk V1.0
Document Version: 1.0 Status: In Progress Date: 11.08.2025
1.0 Overview & Strategic Vision
1.1. Executive Summary
ComplyArk is a comprehensive, on-premise compliance suite designed from the ground up for India's Digital Personal Data Protection Act (DPDPA). It empowers Indian Data Fiduciaries, from high-growth startups to large enterprises, to automatically discover their personal data, manage consent, and orchestrate compliance workfl ows. Through its dual-application architecture—Ark Data Lens for data intelligence and Ark CMS for consent and rights management—ComplyArk replaces manual, high-risk processes with a secure, auditable, and intuitive system. Its core differentiator is its on-premise deployment model, ensuring that a client's sensitive data never leaves their own infrastructure, thus providing maximum security and control.
1.2. Problem Statement
The enactment of the DPDPA presents Indian organizations with a signifi cant operational and legal challenge. Data Fiduciaries are now accountable for every piece of personal data they process, yet most lack the necessary tools and visibility to meet their obligations effectively. The primary pain points are:
● Data Blindness & Sprawl: Personal data is often scattered across hundreds of databases, applications, and unstructured fi les ("data silos"). Organizations lack a centralized, accurate inventory, making it impossible to know what data they hold, where it is, why they have it, and with whom it is shared.
● Manual Compliance Burden: Fulfi lling Data Principal Requests (DPRs)—such as for access or erasure—is a highly manual, error-prone, and time-consuming process. It involves inter-departmental coordination via emails and spreadsheets, which is ineffi cient and leaves no verifi able audit trail.
● Signifi cant Compliance & Financial Risk: The DPDPA carries substantial penalties for non-compliance. Without a systematic way to manage consent, enforce data retention policies, and demonstrate fulfi llment of DPRs, organizations are exposed to signifi cant legal and fi nancial risk.
● Legacy Data & Consent Debt: Organizations hold vast amounts of "legacy data" collected under previous, less stringent consent models. Re-validating this consent to meet the DPDPA's specifi c requirements is a massive, one-time project that is nearly impossible to manage manually.
1.3. Product Vision
To become the defi nitive DPDPA compliance platform for India, empowering every organization to transform their data protection obligations from a source of risk into a demonstration of trust. We envision a future where ComplyArk's integrated suite makes demonstrating compliance a simple, automated, and continuous process, freeing our clients to focus on their core business with confi dence.
1.4. Business Goals & Objectives for V1.0
The release of ComplyArk V1.0 is intended to achieve the following strategic business objectives:
1. Establish Market Leadership: Launch a market-ready Minimum Viable Product (MVP) that addresses the most critical and urgent DPDPA obligations, positioning ComplyArk as the go-to, purpose-built solution.
2. Validate Product-Market Fit: Successfully deploy ComplyArk V1.0 with a cohort of pilot clients across different industries to prove its real-world value, gather critical feedback, and establish strong case studies.
3. Build a Scalable Foundation: Engineer a robust, modular architecture that not only delivers V1.0 features reliably but also provides a solid foundation for future development of advanced features (e.g., SDF obligations, real-time APIs, AI-driven automation).
1.5. Success Metrics (KPIs)
The success of ComplyArk V1.0 will be measured against the following key performance indicators, which directly refl ect its value to our clients:
● Time-to-Value: A new client can establish their baseline compliance posture—defi ned as completing their initial data inventory and launching their fi rst DPDPA-compliant re-consent campaign—within 15 business days of deployment.
● Compliance Effi ciency: ComplyArk reduces the man-hours required to fulfi ll a complex, multi-system Data Principal Erasure Request by over 85% compared to a fully manual process.
● Audit Readiness: 100% of all auditable compliance actions (consent changes, DPR submissions and resolutions, breach notifi cations) generate a corresponding, immutable entry in the audit log without requiring any manual logging by the user.
● User Adoption: Within the fi rst three months of deployment, over 80% of a client's internal departments that process personal data are onboarded and actively using the Department Dashboard to manage their compliance tasks (e.g., resolving child tickets for DPRs).
2.0 User Personas & Roles
2.0.1 Introduction
The following personas represent the key individuals who will interact with the ComplyArk suite. Every feature, user interface element, and workfl ow described in this document is designed to address the specifi c goals and alleviate the frustrations of these users. They are the foundation of our user-centric design philosophy, ensuring we build a product that is not only powerful but also intuitive and effective in the real world.
2.1 Persona 1: David, the Data Fiduciary Administrator (The "Guardian")
● Role & Responsibilities: David is the technical backbone of the organization's IT infrastructure. He is responsible for deploying, maintaining, and securing all on-premise applications, including ComplyArk. His daily tasks involve managing server health, applying updates, confi guring user access according to company policy, and ensuring all systems are integrated and running smoothly. He is the fi rst line of defense and support for any technical issues related to the platform.
● Goals: David's primary goal is system stability. He seeks a "set-it-and-forget-it" platform that runs reliably without constant intervention. He values a seamless, well-documented installation and update process that can be scripted and automated. For him, granular control over user permissions is non-negotiable, as it's his job to
prevent unauthorized access and enforce the principle of least privilege. Clear performance dashboards are essential for him to monitor system load and proactively manage resources.
● Frustrations: David is frustrated by "black box" applications that offer poor logging and make debugging a nightmare. He is deeply concerned by any software that could introduce security vulnerabilities into his carefully managed infrastructure. He dislikes tools that require repetitive manual confi guration instead of offering automated solutions and is particularly irritated by vague error messages that provide no actionable path to resolution.
● How ComplyArk V1.0 Helps David:
○ Simplifi ed Deployment: The Docker-containerized deployment model drastically simplifi es the installation and update process, making it predictable and manageable.
○ Centralized Control: The Data Fiduciary Administrator Dashboard is his command center, providing all necessary controls for SMTP, OTP, and internal user management in one place.
○ Secure by Design: The robust Role-Based Access Control (RBAC) framework allows him to confi dently assign permissions, knowing that users can only access what they are authorized to see and do.
○ Transparent Operations: Detailed system and exception logs provide the transparency he needs to troubleshoot issues quickly and effectively.
2.2 Persona 2: Priya, the Data Protection Offi cer (The "Orchestrator")
● Role & Responsibilities: Priya holds the ultimate responsibility for the organization's DPDPA compliance posture. She orchestrates the entire data protection strategy, from creating and maintaining the Record of Processing Activities (RoPA) to managing the full lifecycle of Data Principal Requests (DPRs) and Grievances. She leads the response to data breaches and is tasked with defi ning and communicating data retention policies. Priya serves as the critical bridge between the legal, technical, and business teams.
● Goals: Priya’s primary objective is to have a "single pane of glass" view of the organization's data risk and compliance status. She needs the ability to verifi ably prove compliance to auditors and regulators at a moment's notice. A core goal is to ensure all DPRs are fulfi lled within the strict timelines mandated by the DPA, and to proactively identify and mitigate privacy risks before they escalate into incidents.
● Frustrations: Priya is constantly frustrated by having to manually chase department heads for information about their data processing activities. Using spreadsheets and email chains to track complex, multi-stage DPRs is her biggest pain point, as it is ineffi cient and prone to human error. She is frequently unable to get a quick, accurate answer to the fundamental question, "Where is this individual's data located across all our systems?"
● How ComplyArk V1.0 Helps Priya:
○ Complete Visibility: The Ark Data Lens provides the comprehensive, centralized data inventory she has always needed, eliminating data silos.
○ Automated Workfl ows: The Ark CMS replaces her manual spreadsheet system with an automated, auditable parent-child ticketing workfl ow for managing DPRs, ensuring nothing is missed.
○ Actionable Insights: The DPO Dashboard provides the "single pane of glass" she requires, with real-time compliance metrics and KPIs.
○ Crisis Management: The Breach Notifi cation module equips her with a tool for rapid, documented communication during a data breach, helping her meet her legal obligations under pressure.
2.3 Persona 3: Mohan, the Department User (The "Enabler")
● Role & Responsibilities: Mohan is a results-driven business leader, such as the Head of Marketing or HR. He and his team process personal data as a core part of their daily operations. His responsibility within the compliance framework is to act promptly on tasks assigned to his department, such as erasing a user's data from their marketing automation platform. He also needs to ensure his team's new initiatives (e.g., launching a new marketing campaign) are compliant from the start.
● Goals: Mohan's goal is to achieve his business objectives with maximum effi ciency. He needs clear, simple rules about what data his team is permitted to use and for what purpose. He wants to spend as little time as possible on compliance-related tasks so he can focus on his primary responsibilities.
● Frustrations: Mohan views compliance as a "black box" that often blocks his team's work without providing clear, actionable reasons. He dislikes receiving vague instructions from the DPO via email and worries that someone on his team could accidentally cause a compliance violation due to a lack of clarity.
● How ComplyArk V1.0 Helps Mohan:
○ Simplicity and Focus: The dedicated Department Dashboard shows him only the tasks and information relevant to his team, eliminating clutter and confusion.
○ Actionable Tasks: The parent-child ticketing system provides him with clear, self-contained tasks ("Erase data for user X from System Y") with all the necessary context.
○ Proactive Compliance: The manual Consent Validation Console empowers him to proactively check the validity of a user list for a new campaign, reducing the risk of non-compliance before he even starts.
2.4 Persona 4: Anjali, the Data Principal (The "Owner")
● Role & Responsibilities: Anjali is a customer of the Data Fiduciary. As the individual to whom the personal data relates, she is the ultimate "owner" of her data under the DPDPA.
● Goals: Anjali wants to interact with services she trusts. She seeks to easily understand how her personal data is being used and wants simple, accessible controls to manage her consent preferences at a granular level. When she chooses to exercise her data rights (like access or erasure), she expects a straightforward process and a timely, clear response.
● Frustrations: Anjali is frustrated by long, convoluted privacy policies and websites where privacy settings are deliberately hard to fi nd. She dislikes "all-or-nothing" consent forms that don't allow her to make specifi c choices. Her biggest frustration is submitting a data request and either never hearing back or receiving a confusing, unhelpful automated reply.
● How ComplyArk V1.0 Helps Anjali:
○ Empowerment through a Portal: The external-facing Data Principal Portal, powered by Ark CMS, gives her a single, secure, and user-friendly destination for all her privacy-related interactions with the company.
○ Granular Control: The portal's UI is designed to allow for easy, granular modifi cation and withdrawal of consent for specifi c purposes.
○ Transparency and Trust: The DPR submission form is simple, and the system provides her with a reference number and real-time status updates, building trust through transparency.
2.5 Persona 5: The SuperAdmin (The "Architect")
● Role & Responsibilities: The SuperAdmin is a skilled engineer or technical consultant from the ComplyArk team. Their role is to architect the solution for the client, handling the initial deployment, setup, and confi guration of ComplyArk on the client's infrastructure. They perform scheduled maintenance, apply version updates, and are responsible for implementing client-specifi c customizations. They are the ultimate expert on the platform, troubleshooting the most complex issues.
● Goals: The SuperAdmin's primary goal is to execute a quick, successful, and repeatable deployment process. They need a powerful set of tools to diagnose and resolve client issues effi ciently, ideally without requiring direct server access for every minor problem. They must be able to manage multiple client instances and their unique confi gurations in a scalable manner.
● Frustrations: The SuperAdmin is most frustrated by deployment processes that fail due to unforeseen client environmental factors. A lack of a centralized view of all managed client instances makes their job diffi cult. They are wary of client customizations that are hard to manage and create divergent, un-maintainable branches of the product.
● How ComplyArk V1.0 Helps the SuperAdmin:
○ Dedicated Toolset: A dedicated SuperAdmin dashboard will provide the necessary tools for client instance management, feature fl agging, and confi guring master role templates.
○ Enhanced Diagnostics: Robust, centralized application logging, which can be securely accessed, will allow for more effi cient remote diagnostics.
○ Scalable Architecture: The platform's architecture is designed to separate core product logic from client-specifi c confi gurations, making maintenance and updates more manageable across the entire client base.
3.0 Key End-to-End User Journeys (V1.0)
3.0.1 Introduction
These user journeys are practical, narrative-driven scenarios illustrating how ComplyArk's personas will interact with the platform to achieve high-value, compliance-critical goals. They serve to demonstrate the interconnectedness of the modules and translate abstract feature requirements into tangible workfl ows, ensuring that the end product is both logical and intuitive. These aren't exhaustive list of all user fl ows
3.1 Journey 1: Establishing the Initial Compliance Baseline ("Day Zero" for a New Client)
● Goal: To guide a newly onboarded organization from a state of data ambiguity to having a foundational, DPDPA-ready data inventory and launching its fi rst re-consent campaign. This journey is the critical fi rst step in achieving auditable compliance.
● Personas Involved: David (Data Fiduciary Administrator), Priya (Data Protection Offi cer).
● Pre-conditions: ComplyArk V1.0 has been successfully deployed within the client's on-premise or cloud environment. David and Priya have received their login credentials for their respective roles.
● Step-by-Step Narrative:
1. System Setup (David): David performs the fi rst login into the Data Fiduciary Administrator Dashboard. His initial task is to confi gure the system for communication. He navigates to System Administration > SMTP Confi guration and enters the organization's email server details, sending a test email to confi rm the setup is working.
2. Data Source Connection (David): David proceeds to the Ark Data Lens application. Within the Fiduciary Confi guration > Data Connectors module, he clicks "Add New Source." He selects "PostgreSQL" from the dropdown and enters the read-only credentials, host, and port for the company's primary customer database. He repeats this process for the Marketing team's MongoDB instance. The dashboard shows both connections as "Active" after successful testing.
3. Initiate Discovery (Priya): Now that the sources are connected, Priya logs into her DPO dashboard and navigates to the Data Lens. She initiates the fi rst data discovery scan by clicking "Start Scan." The system informs her that the scan will run in the background and she will be notifi ed upon completion (or she can wait for the default nightly scan).
4. Review Discovered Data (Priya): The following morning, Priya reviews the completed scan results in the Data Discovery > Data Inventory module. The inventory is a searchable table listing all discovered databases, schemas, tables, and columns. She observes that the system has automatically fl agged columns named email, phone_number, and pan_card_no with a "PII" tag based on pre-defi ned patterns.
5. Manual Classifi cation & Enrichment (Priya): Priya enhances the inventory with her business knowledge. She selects the last_order_date column and applies a
manual tag of "Transactional Data." She similarly tags the address_line_1 and city columns as "Location Data." This process enriches the raw technical metadata with business-relevant context.
6. Create Data Lineage (Priya): In the Governance > Data Lineage Mapper, Priya uses the simple visual tool. She drags the "Customer DB" icon onto the canvas and then drags the "Marketing Mongo DB" icon. She draws an arrow between them, and a dialog box prompts her to describe the data fl ow. She types, "Customer contact and order history shared for marketing analytics."
7. Defi ne Purposes (Priya): Priya switches to the Ark CMS application, navigating to the DPO Workbench > Notice & Consent Management. Here, she defi nes the specifi c, lawful purposes for data processing, such as P001: Order Fulfi llment & Delivery and P002: Promotional Marketing & Offers.
8. Create Notice (Priya): She creates a new "Legacy Customer Re-Consent Notice." Using a wizard-style interface, she links the data categories she identifi ed in the Data Lens (e.g., "Contact Info," "Location Data") to the purposes she just defi ned (P001, P002). She fi nalizes the clear, plain-language text in English as required by the DPDPA.
9. Translate Notice (Priya): With the English version locked, Priya clicks the "Translate" button. The integrated IndicTrans2 model processes the text and generates versions in the 22 Eighth Schedule languages, each linked to the master English version. Priya spot-checks a few key languages like Hindi and Tamil to ensure accuracy.
10. Launch Re-Consent Campaign (Priya): Finally, Priya navigates to the DF Admin Workbench > Consent Collection module. She creates a "New Bulk Request," selects the "Legacy Customer Re-Consent Notice," and targets the user segment "All Users in Customer DB." She initiates the campaign.
● End State: The system begins dispatching re-consent notices to all legacy customers. Their responses (grant, deny, or modifi cation of preferences) are captured and stored as immutable consent artifacts in Ark CMS. The organization now has a foundational, DPDPA-compliant data inventory and an active process to align its legacy data with the new law.
3.2 Journey 2: Fulfi lling a Data Principal Erasure Request (The "Right to be Forgotten")
● Goal: To verifi ably and effi ciently process a data erasure request from a Data Principal across all relevant systems, creating a complete and defensible audit trail.
● Personas Involved: Anjali (Data Principal), Priya (DPO), Mohan (Department User - Marketing).
● Pre-conditions: Anjali is an existing customer whose data resides in both the production database and the marketing database, as mapped in Data Lens.
● Step-by-Step Narrative:
1. Request Submission (Anjali): Anjali logs into the company's Data Principal Portal (powered by Ark CMS). She navigates to the "My Data Rights" section, selects "Request Data Erasure," and submits the form. The interface immediately provides her with a confi rmation: "Your request has been submitted. Your reference ID is DPREQ-2025-00123."
2. Parent Ticket Creation (System): Instantly, Ark CMS creates a "Parent Ticket" for DPREQ-2025-00123, assigning it to the DPO user group. Priya receives an email notifi cation: "New Erasure Request Received."
3. Child Ticket Creation (System): The system's workfl ow engine automatically queries the Data Lens inventory using Anjali's unique identifi er. The query returns two locations for her PII: the "customer_db" (owned by the IT department) and the "marketing_mongo_db" (owned by the Marketing department). The system then automatically generates two "Child Tickets" linked to the parent ticket: CTICK-00245 is assigned to the IT department queue, and CTICK-00246 is assigned directly to Mohan's Marketing department queue. Mohan receives an email notifi cation.
4. Action & Resolution (Mohan): Mohan logs into his focused Department Dashboard. He sees the new child ticket with the clear instruction: "Erase data for Anjali Sharma (ID: anjali.s@email.com) from Marketing Systems." He proceeds to log into his team's marketing platform and permanently deletes her profi le.
5. Ticket Closure (Mohan): After completing the task, Mohan returns to the Ark dashboard, opens the child ticket, and marks it as "Resolved." He adds a mandatory comment: "User record and all associated campaign history deleted from Marketo instance ID 789."
6. Central Review & Closure (Priya): In her DPO Workbench, Priya monitors the parent ticket. She sees the status of Mohan's child ticket has changed to "Resolved" and can review his comment. The IT department has also resolved their ticket. With all associated child tickets closed, the "Resolve Parent Ticket" button becomes active. Priya performs a fi nal review and clicks it.
7. Final Confi rmation (System): Upon Priya's action, the system automatically triggers a fi nal confi rmation email to Anjali: "Your data erasure request DPREQ-2025-00123 has been completed."
● End State: Anjali's right to erasure has been honored. The entire end-to-end process is captured in a single, auditable record in Ark CMS, detailing every action, owner, and timestamp, providing clear evidence of compliance.
3.3 Journey 3: Compliant Use of Data for a New Business Activity (A "Marketing Campaign")
● Goal: To enable a business team to confi dently use personal data for a new processing activity in a manner that is fully compliant with the consent on record.
● Personas Involved: Mohan (Department User - Marketing), Priya (DPO).
● Pre-conditions: The company wants to run a Diwali promotion targeted at customers in Mumbai who have given consent for marketing.
● Step-by-Step Narrative:
1. Defi ne a User Segment (Mohan): Mohan's team queries their internal systems to generate a list of 5,000 customer emails for users who live in Mumbai. He knows he cannot use this list directly.
2. Verify Consent (Mohan): Mohan logs into his Ark Department Dashboard and navigates to the Tools > Consent Validation console. He uploads the CSV fi le containing the 5,000 emails and selects the specifi c purpose P002: Promotional Marketing & Offers from a dropdown. He clicks "Validate."
3. Receive Validation Report (System): Ark CMS processes the list in seconds, checking each user's latest consent artifact against the selected purpose. The UI returns a clear report: Total Records Submitted: 5000. Valid for Purpose P002: 4250. Invalid (Consent Denied or Withdrawn): 750.
4. Act on Compliant Data (Mohan): Mohan clicks "Download Valid List." He receives a CSV fi le containing only the 4,250 emails that are cleared for the promotion. He confi dently hands this compliant list to his team to be used in the campaign, ensuring the 750 who did not consent are excluded.
5. Log the Activity (System): In the background, Ark CMS creates an audit log entry visible to Priya: "User 'Mohan' initiated a bulk validation for 'P002' on 5000 principals, with 4250 valid results on [Timestamp]."
● End State: Mohan successfully executes his campaign while strictly adhering to consent obligations. Priya, the DPO, has passive oversight and a complete audit trail of
this processing activity, demonstrating purpose limitation without having to be an active bottleneck for the marketing team.
3.4 Journey 4: Managing a Data Breach Notifi cation
● Goal: To manage the critical communication and documentation process following a data breach, ensuring adherence to the DPDPA's strict notifi cation requirements.
● Personas Involved: Priya (DPO), David (DF Admin).
● Pre-conditions: David's security information and event management (SIEM) system has alerted him to an unauthorized export of data from a specifi c customer database.
● Step-by-Step Narrative:
1. Initiate Breach Response (Priya): After an emergency meeting confi rms the breach, Priya's fi rst action is to log into the Ark CMS DPO Workbench and navigate to the Compliance > Breach Notifi cation module. She clicks "Create New Breach Incident."
2. Defi ne Affected Users (Priya): The system prompts her to defi ne the scope. Working with David, she selects the compromised database, "Customer_DB_APAC," from a dropdown list populated by Ark Data Lens. The system immediately shows her the number of Data Principals associated with that data source.
3. Draft Notifi cation (Priya): She uses a pre-built "Breach Notifi cation Template" which contains placeholders for the DPDPA's required elements. She fi lls in the details: the nature of the breach (unauthorized data export), the specifi c data categories involved (contact and order information), the likely consequences, and the immediate remedial measures the company has taken (e.g., revoking credentials).
4. Dispatch Notifi cations (Priya): After a fi nal review, she clicks "Dispatch Notifi cations." Ark CMS queues the emails and sends them via the confi gured SMTP server to all affected Data Principals.
5. Document for the Board (Priya): The module provides a "Generate Incident Report" button. This creates a PDF summary containing the incident timeline, the number of users notifi ed, and the exact text of the notice sent. Priya uses this report as a key document for her formal intimation to the Data Protection Board of India.
● End State: The organization has met its initial legal obligation to inform affected users without delay. The entire communication process is centralized and documented within
ComplyArk, providing a clear, auditable record of the actions taken during a critical security incident.
4.0 Detailed Functional Requirements (V1.0)
4.1. Core Platform & SuperAdmin Module
This section defi nes the foundational components of the ComplyArk suite. These features underpin the entire system, enabling secure administration, multi-tenancy (in a logical sense for the SuperAdmin), and a smooth onboarding process for new clients.
4.1.1. SuperAdmin Authentication & Access
● Description: The SuperAdmin role is reserved exclusively for the ComplyArk technical team. It provides the necessary high-level access to manage and maintain all client instances. Access must be secure, controlled, and auditable.
● User Story:
1. As a SuperAdmin, I need to securely log in to a central management console so that I can view and manage all client deployments.
2. As a SuperAdmin, I need a secure, documented method to gain temporary access to a client's on-premise instance so that I can perform maintenance or troubleshoot critical issues.
● Functional Requirements:
1. Central Login Portal: A dedicated, secure web portal (not accessible from client instances) shall be created for SuperAdmin users.
2. Authentication: SuperAdmin authentication shall use a robust mechanism, such as Multi-Factor Authentication (MFA), in addition to a username and strong password.
3. Secure Instance Access Protocol:
■ The primary method for accessing a client's on-premise instance for maintenance shall be via temporary, client-provided credentials (e.g., VPN access, temporary SSH keys).
■ The product will NOT contain any "backdoor" or "phone-home" mechanism.
■ All SuperAdmin access to a client instance must be logged in the client's system audit log with a clear identifi er (e.g., "Action performed by SuperAdmin [AdminName]").
4. Session Management: SuperAdmin sessions shall have a strict inactivity timeout.
● SuperAdmin Action Logging
● Description: To maintain full transparency and accountability with the client, every action performed by a SuperAdmin user while accessing a client's on-premise instance must be distinctly logged.
● Functional Requirement:: The system shall maintain a dedicated "SuperAdmin Action Log" that is visible to the client's Data Fiduciary Administrator. Each entry in this log must clearly identify the SuperAdmin user, the timestamp of the action, the source IP, and a plain-language description of the action performed (e.g., "Viewed SMTP confi guration," "Triggered a manual database scan"). This ensures that all high-privilege activities are auditable by the client.
4.1.2. Client Instance Management (SuperAdmin UI)
● Description: The SuperAdmin requires a centralized dashboard to effi ciently monitor and manage the fl eet of deployed client instances. This is the SuperAdmin's primary interface for oversight.
● User Story:
1. As a SuperAdmin, I need a dashboard that lists all my clients so that I can quickly assess the status and version of each deployment.
2. As a SuperAdmin, I need the ability to enable or disable specifi c, non-core features for a client so that I can manage custom requirements or phased rollouts.
● Functional Requirements:
1. Instance Dashboard: The SuperAdmin portal shall feature a dashboard displaying a table of all managed client instances.
2. Dashboard Columns: The table shall include the following columns: Client Name, Instance ID, Deployed Version, Instance Status (e.g., Online, Offl ine, Needs Update), Last Heartbeat (if a passive monitoring agent is implemented in the future, otherwise Last Maintained), and Notes.
3. Feature Flag Management: The UI shall allow a SuperAdmin to view and toggle feature fl ags for each client instance. For V1.0, this is a placeholder for future custom features. The system architecture must support this capability from day one.
■ Example: A fl ag enable_advanced_reporting could be toggled on for a specifi c client.
● Master Confi guration Management
1. Description: To ensure consistency and enable systematic improvements across all client deployments, the SuperAdmin requires a tool to manage master confi guration libraries that are packaged with the application.
2. Functional Requirement: The SuperAdmin portal shall include a "Master Libraries" section. For V1.0, this will allow the SuperAdmin to view and manage the master library of PII detection patterns (regex). Future versions will extend this to manage default email templates and other global assets. This library is packaged within new releases, providing an updated baseline for all clients upon updating their ComplyArk instance.
4.1.3. Role & Permission Management (SuperAdmin UI)
● Description: To ensure scalability and consistency, the SuperAdmin must be able to defi ne the master templates for user roles and permissions that will be deployed to all client instances.
● User Story:
1. As a SuperAdmin, I need to defi ne the default permissions for the "DPO" role so that every new client instance starts with a secure and correct permission set.
2. As a SuperAdmin, I need to be able to update these role templates and push them out with new releases so that I can introduce new features and permissions systematically.
● Functional Requirements:
1. Master Role Template Editor: The SuperAdmin portal shall include a UI for managing master role templates (Data Fiduciary Administrator, DPO, Department User).
2. Permission Granularity: The editor shall display a checklist of all available system permissions (e.g., view_data_inventory, create_notice, resolve_parent_ticket, confi gure_smtp). The SuperAdmin can select which permissions apply to each master role.
3. Deployment Mechanism: The defi ned master role templates (Data Fiduciary Administrator, DPO, Department User) shall be packaged as a core confi guration fi le within the ComplyArk Docker container. When a new instance is deployed or an existing one is updated, it will use these master templates to populate its internal Role-Based Access Control (RBAC) system. The client-side DF Admin can then create users and assign them these pre-defi ned roles.
**Crucially, for V1.0, the DF Admin cannot alter the fundamental permissions of these master roles.** This maintains a secure, consistent baseline across all deployments and prevents permission misconfi gurations.
4.
4.1.4. Onboarding & Setup (First-Run Experience)
● Description: The initial interaction a client's DF Admin has with ComplyArk is critical. This workfl ow must be a simple, guided experience that sets up the foundational confi guration for their instance.
● User Story:
1. As a DF Admin logging in for the fi rst time, I need a setup wizard to guide me through the essential confi guration steps so that I can get the system operational quickly.
2. As a DF Admin, I need to fi ll out my organization's details once so that they can be automatically used in reports and notices.
● Functional Requirements:
1. First-Run Detection: The system shall detect when a DF Admin is logging in for the fi rst time and automatically launch a mandatory setup wizard.
2. Onboarding Questionnaire: The wizard shall present a multi-step form to capture:
■ Organization Details: Company Legal Name, Address, Corporate Identity Number (CIN). This data will be used as variables in reports.
■ Initial User Creation: A form to create the primary DPO user account (Name, Email).
3. Initial Department Setup: The wizard will prompt the DF Admin to defi ne their organization's internal departments that handle personal data (e.g., Marketing, Human Resources, IT Support). This will be a simple interface to add/edit/delete department names, which are then used for assigning Department Users and for creating the Data Lineage Map.
4.
5. Initial Confi guration: The wizard will guide the DF Admin to the System Administration > SMTP Confi guration page to ensure email notifi cations are functional from the start.
6. Wizard Completion: Upon completing the wizard, the DF Admin will be directed to the main application dashboard, and the wizard will not appear on subsequent logins.
4.2. Ark Data Lens: Functional Breakdown
4.2.0 Introduction
Ark Data Lens serves as the foundational "System of Record" for an organization's personal data landscape. Its primary function is to empower Data Fiduciaries to move from a state of data ambiguity to one of complete visibility. The modules within Data Lens are designed to systematically answer the fundamental questions of DPDPA compliance: what personal data do we hold, where is it stored, what is it used for, and how does it move through our systems? The output of Data Lens is the critical intelligence that fuels the compliance workfl ows of Ark CMS.
4.2.1. Module: Data Source Connector Management
● Description: This module provides the interface for the Data Fiduciary Administrator to securely connect Ark Data Lens to the organization's various data repositories. It is the gateway through which the application gains the necessary (read-only) visibility.
● User Story: As a DF Admin (David), I need to easily add, confi gure, and monitor connections to all our databases and fi le stores so that the Data Discovery engine has a complete and accurate view of our data landscape.
● Functional Requirements:
1. Connections Dashboard: The module's main page shall display a table of all confi gured data source connections with the following columns:
■ Connection Name: A user-defi ned alias (e.g., "Production Customer DB").
■ Source Type: An icon/label for the technology (e.g., PostgreSQL, S3).
■ Status: A visual indicator (e.g., green dot for "Connected," red for "Error," grey for "Disabled").
■ Last Scanned On: Timestamp of the last successful scan.
2. Add/Edit Connection Workfl ow:
■ Clicking "Add New Source" shall open a modal or dedicated page.
■ A mandatory dropdown fi eld, Source Type, will dictate the subsequent form fi elds.
■ For Database Types (PostgreSQL, MySQL, MS SQL, MongoDB, Oracle): The form shall require: Connection Name, Hostname/IP
Address, Port, Database Name, Read-Only Username, and Read-Only Password (fi eld must be masked).
■ Security Mandate: All submitted credentials (passwords, secret keys) must be immediately encrypted using a strong, industry-standard algorithm (e.g., AES-256 with a managed key) before being stored in the ComplyArk database. The raw credentials shall never be stored in plain text and shall not be viewable in the UI after the initial entry.
■ For S3 Buckets: The form shall require: Connection Name, Bucket Name, AWS Region, Read-Only Access Key ID, and Read-Only Secret Access Key (masked).
■ For Local File System: The form shall require: Connection Name and the absolute File System Path (this path must be mounted and accessible to the ComplyArk Docker container).
■ Connection Testing: A "Test Connection" button is mandatory. It must provide immediate, clear feedback (e.g., "Connection successful!" or "Error: Authentication failed. Please check credentials.") before the "Save" button is enabled.
3. Connection Management: From the dashboard, an admin must be able to Edit (re-open the confi guration form), Disable/Enable (toggle a connection's active status for scans), or Delete a connection (with a confi rmation prompt).
4.2.2. Module: Data Discovery & Inventory
● Description: This is the core engine of Data Lens. It performs the scans and aggregates all discovered metadata into a centralized, searchable inventory.
● User Story: As a DPO (Priya), I need the system to automatically scan our connected sources and present me with a comprehensive inventory of all data assets, so I can begin the process of classifi cation and risk assessment.
● Functional Requirements:
1. Scanning Engine:
■ Scheduler: A settings panel shall allow the DF Admin to confi gure the automatic nightly scan time (defaulting to 2:00 AM local time).
■ Manual Trigger: A "Start Full Scan Now" button shall be prominently displayed on the Lens dashboard for on-demand discovery, with a visual indicator showing the scan is in progress.
■ Scan Logic: The scanner will ingest metadata only. This includes database schemas, table/view names, column names, data types, folder structures, fi le names, and fi le extensions. It will not read or copy the data content itself.
■ Scan History: A dedicated log page will show a reverse chronological list of all scans with Start Time, End Time, Status (Completed, Failed, In Progress), and a summary (e.g., "Discovered 5 new tables, 30 new columns").
2. Automated PII Pattern Recognition (V1.0 - Suggestion Only):
■ The system will contain a built-in, non-editable library of regex patterns for common Indian PII.
■ During a scan, if a column name itself matches a pattern (e.g., user_pan, aadhar_no), the system will apply a "System Suggested: PII" tag to that asset in the inventory. This is for guidance and must be manually confi rmed.
3. The Data Inventory UI:
■ This UI is a powerful, interactive table serving as the master list of all data assets.
■ Filtering & Searching: The UI must include powerful, multi-select fi lters for Source Name, Asset Type (e.g., Table, Column, File), Tags, and Classifi cation Status (Classifi ed/Unclassifi ed). A global search bar will search across all metadata fi elds.
■ Drill-Down View: The inventory will be structured hierarchically. Clicking on a Source will fi lter the view to its Tables/Folders; clicking a Table will show its Columns.
4. Manual Asset Registration
■ Description:** To ensure the data inventory can be truly comprehensive, the system must account for data that exists in systems that cannot be connected to electronically.
■ Functional Requirement: The Data Inventory UI shall include a "Register Manual Asset" function. This will open a form where a DPO or DF Admin can document offl ine or disconnected data stores. The form will require: Asset Name (e.g., "Employee Contracts Filing Cabinet"), Physical Location (e.g., "Mumbai Offi ce, HR Department"), Data Categories Contained (multi-select from tags), and Responsible Department. These manual assets will appear in the inventory alongside discovered digital assets, allowing them to be included in the holistic data map.
4.2.3. Module: Governance & Mapping
● Description: This module provides the tools for the DPO to enrich the raw inventory with business context, turning technical metadata into a meaningful compliance map.
● User Story: As a DPO (Priya), I need tools to manually classify our data assets, link them to business purposes, and visually map how data fl ows between systems so I can create an accurate Record of Processing Activities (RoPA).
● Functional Requirements:
1. Tagging & Classifi cation System:
■ A settings page will allow the DPO to create, edit, and manage a library of custom tags (e.g., "Employee PII," "Customer Financials") with descriptions.
■ From the Data Inventory UI, the DPO can select one or more data assets (columns, fi les) and apply tags from the library. This manual action formally classifi es the data.
2. Data Lineage Mapper (Manual V1.0):
■ The module will feature a clean, intuitive drag-and-drop canvas.
■ A palette of nodes (Data Source, Department, Processor/Third-Party) will be available.
■ Users can drag nodes onto the canvas, name them (e.g., a "Processor" node named "AWS S3 - Mumbai"), and arrange them visually.
■ Users can draw directed arrows between nodes to represent data fl ow. On creating a fl ow, a dialog box will prompt for details: Flow Description, Data Categories Transferred (multi-select from tags), and Transfer Method (e.g., API, SFTP).
■ The entire map can be saved as a versioned "Data Map."
3. Policy Association:
■ A central library allows the DPO to defi ne Data Retention Policies (Policy Name, Duration, Description).
■ In the Data Inventory, the DPO can associate data assets with a defi ned policy. The Retention Policy column in the inventory will display the name of the associated policy.
4. Map Versioning: The system must support versioning for the Data Lineage Map. A "Save as New Version" button will be present, allowing the DPO to create a historical snapshot of the map before making signifi cant changes. A "Version History" dropdown will allow the DPO to view and compare previous
versions of the map, which is critical for demonstrating the state of data fl ows at a specifi c point in time during an audit.
4.2.4. Module: Data Understanding Dashboard (The Lens Homepage)
● Description: This is the main landing page for Ark Data Lens, providing a high-level, at-a-glance overview of the organization's data landscape and compliance mapping progress.
● User Story: As a DPO (Priya), when I log in, I want to see a summary of our data environment and key risk indicators so I can quickly identify areas that need my attention.
● Functional Requirements:
1. At-a-Glance Widgets: The dashboard will feature several key metric cards:
■ Total Data Sources: A count of all active connections.
■ PII Elements Discovered: A count of all assets tagged as PII (system-suggested or manual).
■ Asset Classifi cation Progress: A progress bar/percentage showing (Manually Tagged Assets / Total Assets) * 100.
■ Unmapped Data Sources: A count of connected sources that do not yet appear on the Data Lineage Map.
2. Interactive Map Visualization:
■ The dashboard will display a prominent, read-only, embedded view of the latest saved Data Lineage Map.
■ Hovering over a node or fl ow on the map will display its name and description in a tooltip.
3. Reporting & Exporting:
■ A "Generate Report" function will be available.
■ Data Inventory Report: This option will export the current view of the Data Inventory table (respecting any active fi lters) to a formatted CSV fi le.
■ Record of Processing Activities (RoPA) Report: This option will generate a comprehensive PDF document. The report will be structured according to DPDPA requirements, programmatically combining data from the Inventory (data categories), the Purpose Library (from CMS), and the Lineage Map (processors, transfers) into a formal, presentation-ready document.
4.3. Ark CMS: Functional Breakdown
4.3.0 Introduction
Ark CMS is the operational core of the ComplyArk suite, designed as the "System of Action and Audit." It translates the data intelligence gathered by Ark Data Lens into tangible, auditable compliance workfl ows. This system enables organizations to manage the entire consent lifecycle, fulfi ll Data Principal rights in a structured manner, and document every compliance-related activity immutably, ensuring they can confi dently demonstrate adherence to the DPDPA.
4.3.1. Module: Data Principal Portal
● Description: This is a secure, external-facing web application that serves as the single point of contact for Data Principals to manage their privacy preferences and exercise their rights with the Data Fiduciary. Its design must be simple, intuitive, and accessible, requiring no technical expertise from the end-user.
● User Story: As a customer (Anjali), I want a single, easy-to-fi nd place where I can see exactly what data the company has, what they are using it for, and have simple controls to change my mind or ask for my data to be deleted.
● Functional Requirements: 1. Secure Authentication:
○ The portal shall be accessible via a unique, stable URL (e.g., privacy.company.com) provided by the Data Fiduciary in their privacy notice and other communications.
○ The login page shall require a unique identifi er. For V1.0, this will be the Data Principal's registered email address or mobile number.
○ Authentication will be performed via a time-sensitive One-Time Password (OTP) sent to the provided identifi er. The OTP shall expire in 10 minutes and be limited to 5 failed attempts before a temporary lockout. The DF Admin will confi gure these parameters.
○ All user sessions will have a strict inactivity timeout (confi gurable by the DF Admin, defaulting to 30 minutes) to ensure security.
● 2. "My Consents" Dashboard:
○ This is the main landing page after successful login. It must provide a clear, at-a-glance view of the user's consent status.
○ The dashboard shall display a list of all processing purposes for which consent has been requested. Each item in the list will represent a distinct consent artifact and must display:
■ Purpose of Processing: A clear, plain-language title (e.g., "Product Recommendations," "Promotional Newsletters").
■ Date of Action: The timestamp of when the consent was last given or withdrawn.
■ Status: A clear visual label, such as "Active" (green), "Withdrawn" (red), or "Expired" (grey).
○ Detailed View: Clicking "View Details" on any purpose shall expand the item to show:
■ The specifi c categories of personal data being processed for that purpose (e.g., "Contact Information," "Purchase History").
■ A direct, downloadable link to the exact version of the Privacy Notice that was in effect when they provided consent.
○ Consent Modifi cation:
■ For each "Active" consent, a "Withdraw Consent" button must be present.
■ Clicking "Withdraw Consent" shall trigger a confi rmation modal (e.g., "Are you sure you want to withdraw consent for Promotional Newsletters? This cannot be undone.").
■ Upon confi rmation, the system shall update the consent artifact status to "Withdrawn" in real-time, and the action will be recorded in the audit log. The UI must immediately refl ect this change.
● 3. Data Principal Rights (DPR) & Grievance Forms:
○ A dedicated "Exercise Your Rights" or "My Requests" section shall be clearly accessible from the main navigation.
○ This section will contain simple, individual web forms for each type of request:
■ Request for Access: A simple submission form. The output for the user will be a summary of their data.
■ Request for Correction: A form with a text area where the user can describe the inaccurate data and provide the correct information.
■ Request for Erasure: A simple submission form with a confi rmation step.
■ Request for Nomination: A form requiring the nominee's name, email, and relationship, along with the user's confi rmation.
■ Lodge a Grievance: A form with a dropdown for the grievance category, a text area for a detailed description, and an optional fi le upload fi eld (max 5MB, limited to common fi le types like PDF, JPG, PNG).
○ All forms, upon submission, will provide the user with an on-screen confi rmation message and a unique, trackable Request ID.
● 4. Request Tracking Dashboard:
○ This page shall display a table of all DPRs and Grievances submitted by the logged-in Data Principal.
○ The table must include the following columns: Request ID, Request Type (e.g., Erasure, Grievance), Date Submitted, and Current Status.
○ The Status fi eld will be a real-time refl ection of the request's state in the DPO's workfl ow engine (e.g., "Submitted," "In Progress," "Resolved," "Closed").
○ Clicking on a request ID will show a detailed history and any comments or updates provided by the DPO.
4.3.2. Module: DPO & Compliance Workbench
● Description: This workbench is the DPO's command and control center within Ark CMS. It is an internal-facing interface designed to consolidate all critical compliance tasks, monitoring dashboards, and workfl ow management into a single, cohesive environment. This module empowers the DPO to move from a reactive to a proactive compliance stance.
● User Story: As a DPO (Priya), I need a centralized workbench where I can instantly see my operational workload, manage data rights requests from start to fi nish, and access all the tools I need to oversee our organization's DPDPA compliance.
● Functional Requirements: 1. Central DPO Dashboard (The Homepage):
○ This is the default landing page for any user with the DPO role. It is designed to provide an immediate, at-a-glance summary of the organization's compliance health and operational priorities.
○ Key Metric Widgets: The dashboard will feature a series of non-confi gurable, real-time metric widgets:
■ Open DPRs: A live numerical count of all Data Principal Requests with a status of "Submitted" or "In Progress." Clicking this widget navigates the user to the DPR queue, pre-fi ltered for these statuses.
■ Overdue DPRs: A prominent, high-visibility (e.g., red-colored) count of DPRs that have breached the SLA confi gured by the DF Admin. Clicking this navigates to the DPR queue, pre-fi ltered to show only overdue requests.
■ Pending Parental Consents: A count of consent requests from minors or persons with disabilities that are awaiting DPO review and approval.
■ Open Grievances: A live count of all unresolved grievance tickets.
■ Recent High-Priority Audit Events: A feed displaying the last fi ve critical events from the system's audit log, such as "Consent Withdrawn," "Breach Incident Created," "SuperAdmin Login," or "Notice Published."
● 2. DPR & Grievance Workfl ow Engine:
○ This is the core operational tool for managing data rights.
○ Master Queue Interface:
■ A unifi ed table shall display all incoming Parent Tickets for both DPRs and Grievances.
■ The table must contain the following sortable columns: Ticket ID, Data Principal Identifi er (e.g., user@email.com), Request Type (e.g., Erasure, Access, Grievance), Date Submitted, SLA Timer (a visual countdown, e.g., "28 days remaining," which turns red when overdue), and Status (e.g., Submitted, In Progress, Resolved).
■ The queue must have powerful header fi lters, allowing the DPO to fi lter the view by any column value (e.g., show only "Erasure" requests).
○ Parent Ticket Detailed View:
■ Clicking a Ticket ID shall open a dedicated page for that request.
■ The view will be logically divided:
■ Header Section: Displays the full details of the Data Principal's original request, their contact information, and the parent ticket's overall status and SLA.
■ Child Tickets Section: A list of all system-generated Child Tickets linked to this request. Each child ticket will show its Child Ticket ID, the Assigned Department (from Data Lens mapping), and its real-time Status (Pending, Resolved).
■ Internal Comments & History: A chronological log and comment thread. The DPO and assigned department users can add timestamped comments. All system actions (e.g., "Child ticket created and assigned to Marketing") are automatically logged here.
■ Actions Panel (DPO only): The DPO will have buttons to Add Comment or Re-assign Ticket to another DPO user. The master Resolve Parent Ticket button will be disabled by default and will only become enabled once the status of all associated Child Tickets is "Resolved."
○ Child Ticket Management (Accessed via Department Dashboard):
■ Department Users (Mohan) will see a simplifi ed queue of only the child tickets assigned to their department.
■ The actions available to them are Add Comment and Mark as Resolved. Clicking "Mark as Resolved" will prompt for a mandatory resolution comment (e.g., "User's data has been purged from our Salesforce instance.").
● 3. Verifi able Parental Consent Queue:
○ A dedicated section within the DPO workbench, labeled "Parental Consent Approvals," will list all pending consent requests from minors or persons with disabilities.
○ Review Interface: Clicking on a request will open a detailed view showing the minor's submitted information, the guardian's submitted information, and the Verifi cation Method used.
○ Document Verifi cation: If the verifi cation method was "Manual Upload," the interface must securely render the uploaded ID proof documents within the browser for the DPO to review. Direct download will be possible.
○ Decision Workfl ow:
■ The DPO will have two primary actions: Approve and Reject.
■ If Reject is chosen, a modal will appear requiring the DPO to enter a reason for the rejection. This reason is logged and used in the notifi cation sent back to the guardian.
■ If Approve is chosen, the system marks the consent as verifi ed, and the action is recorded in the audit log. This approval can trigger the next step in the client's actual onboarding process via an API webhook (a V1.1 feature, but the approval logic must be present in V1.0).
● 4. Policy & Training Repository:
○ This feature provides a simple, centralized document library for managing key compliance artifacts.
○ Document Library UI: The interface will allow the DPO to upload, view, and manage documents. The view will be a table with columns for Document Name, Version, Owner, and Last Updated Date.
○ Functionality:
■ Upload New Document: Allows the DPO to upload a new policy (PDF, DOCX format).
■ Upload New Version: When viewing an existing document, this option allows the DPO to upload a new fi le, which will be marked as the latest version, while the previous version is archived but remains accessible.
■ Download: Users can download any version of a stored document.
■ This repository is intended to be the single source of truth for internal policies such as the Data Retention Policy, Cybersecurity Incident Response Plan, and Data Processing Agreements (Templates).
● 5. Data Access Request Fulfi llment Workfl ow: Description: The process for fulfi lling a Data Access request must be explicitly defi ned to ensure it is auditable, even though ComplyArk does not access the data itself.
● Functional Requirement: When a DPO is ready to resolve a Data Access request ticket, the "Resolve" action will open a specifi c modal. This modal will require the DPO to **upload the data summary fi le** (e.g., a PDF or CSV) that they have manually compiled from their source systems. Upon successful upload, the DPO can add a resolution comment and close the ticket. The system will log this event and automatically make the uploaded fi le securely available for the Data Principal to download from their portal's request tracking page. This workfl ow ensures ComplyArk orchestrates and documents the delivery without ever handling the raw personal data directly.
4.3.3. Module: Notice & Consent Lifecycle Management
● Description: This module provides the essential tools for the Data Protection Offi cer (DPO) to create, manage, and deploy legally sound notices and consent collection fl ows. It is the bridge between the organization's internal data governance (as mapped in Data Lens) and its external, transparent communication with the Data Principal. This module ensures that consent, the cornerstone of the DPDPA, is handled in a manner that is specifi c, informed, version-controlled, and fully auditable.
● User Story: As a DPO (Priya), I need a systematic way to build DPDPA-compliant notices based on our actual data processing, get them translated, and make them available to our development team for integration, all while maintaining a perfect version history.
● Functional Requirements: 1. Purpose Library:
○ This is a centralized repository for all data processing purposes within the organization, ensuring consistent terminology and preventing "purpose sprawl."
○ UI & Functionality:
■ Located within the DPO Workbench, this feature will present a table with columns: Purpose ID, Purpose Name, Purpose Description, and Status (Active/Inactive).
■ Create Purpose: A "Create New Purpose" button will open a modal requiring:
■ Purpose Name: A short, user-friendly label (e.g., "User Account Management," "Promotional Marketing"). Max 100 characters.
■ Purpose Description: A detailed, plain-language explanation of the processing activity, what it entails, and the value to the Data Principal. Max 500 characters.
■ Upon saving, the system shall generate a unique, non-editable Purpose ID (e.g., PURP-001).
■ Edit/Deactivate: A DPO can Edit the Name and Description of a purpose. A DPO can Deactivate a purpose, which prevents it from being added to new notices but does not affect existing notices where it is already used (to maintain historical integrity).
● 2. Dynamic Notice Generator:
○ This is a wizard-style interface designed to guide the DPO through the creation of a legally robust notice, ensuring all elements of DPDPA Section 5 are met.
○ Notice Creation Workfl ow:
■ Step 1: Basic Details: The DPO initiates a new notice, providing:
■ Notice Name: An internal-facing name (e.g., "E-commerce Customer Onboarding Notice v1.0").
■ Grievance Offi cer Details: Select the DPO or designated Grievance Offi cer from a dropdown of confi gured users. Their contact details will be auto-populated into the notice.
■ Step 2: Link Purposes and Data Categories: This is the core compliance-linking step.
■ The UI will allow the DPO to select one or more purposes from the Purpose Library.
■ For each purpose selected, the UI will require the DPO to link the specifi c Data Categories (these are the tags defi ned in Ark Data
Lens, such as "Contact Information," "Financial KYC Data"). This direct linkage enforces the principle of purpose limitation.
■ Step 3: Draft Content:
■ A rich-text editor is provided for drafting the main body of the notice.
■ The system will automatically generate and insert structured sections based on the previous step, for example: For the purpose of "Order Fulfi llment & Delivery," we will process the following categories of your data:
■ Contact Information (e.g., name, email, phone number)
■ Location Data (e.g., shipping and billing addresses)
■ The DPO can add introductory and concluding text around these auto-generated sections.
■ Step 4: Review & Finalize: A fi nal "Review" screen presents the complete notice text as it will be shown to the Data Principal. The DPO must tick a confi rmation box: "I have reviewed this notice and confi rm its contents are accurate."
■ Clicking "Finalize & Lock Version" saves this notice as v1.0 (or the next incremental version). This action is irreversible. The notice is now ready for translation.
○ Version Control: The system must maintain an immutable version history for every notice. A DPO can never edit a fi nalized version. Instead, they must use a "Create New Version" function, which duplicates the latest version into a new draft for editing. The notice management dashboard will clearly show the full version history (v1.0, v1.1, v2.0), with only one version being designatable as "Active" for new consent fl ows.
● 3. Translation Engine:
○ This feature facilitates compliance with the DPDPA's multi-language requirement.
○ Workfl ow:
■ On a fi nalized and locked notice version, a "Translate" button becomes available.
■ Clicking this button initiates a one-time, irreversible background job for that specifi c version. The system sends the master English text to the integrated IndicTrans2 AI model.
■ The model returns translations for the 22 Eighth Schedule languages, which are then stored and linked to the master version.
○ UI & Management:
■ After the translation job is complete, a new "Translations" tab appears on the notice page.
■ This tab displays a table listing all 22 languages, the status (Completed), and an option to "View."
■ "View" opens a side-by-side comparison of the master English text and the AI-generated translation, allowing for DPO review.
■ A permanent, non-dismissible disclaimer must be displayed prominently on this page: "Disclaimer: The following translations are generated by an automated AI model. ComplyArk does not guarantee their legal or contextual accuracy. It is the Data Fiduciary's sole responsibility to review and verify the correctness of all translations before publication."
● 4. Consent Collection Flow Management:
○ This feature, located in the DF Admin Workbench, is the fi nal step that connects a prepared notice to the client's public-facing applications.
○ UI & Functionality:
■ A UI allows an admin to "Create New Consent Flow."
■ The creation form requires:
■ Flow Name: A descriptive name for the integration point (e.g., "Main Website Signup Flow," "Android App Onboarding").
■ The user must then select one Active, Translated Notice from a dropdown list. This list only shows notices that have been fi nalized and have had the translation process completed.
○ API Key Generation:
■ Upon saving the fl ow, the system generates and displays the following for the client's development team:
■ A unique Flow ID (e.g., fl ow-a1b2c3d4).
■ A secure, randomly generated API Key. The key is shown only once, and the user must copy it.
■ The exact API endpoint URL for this fl ow (e.g., https://client.ark.local/api/v1/consent-fl ow/{Flow ID}).
○ API Specifi cation (to be provided to client developers):
■ The API must be fully documented.
■ GET /api/v1/consent-fl ow/{Flow ID}: This endpoint, when called with a valid API key and a language code (e.g., ?lang=hi), will return the structured JSON content of the appropriate notice version.
■ POST /api/v1/consent-fl ow/{Flow ID}/consent: This endpoint receives the Data Principal's unique identifi er and their consent choices (a list of Purpose IDs they granted). The system validates the submission and creates the immutable consent artifact in the database, returning a 201 OK success response.
4.3.4. Module: Data Fiduciary Admin Workbench
● Description: This workbench serves as the operational hub for the Data Fiduciary Administrator and other authorized power users. While the DPO Workbench focuses on governance and fulfi llment, this module provides the tools to proactively validate data usage, manage third-party relationships, and monitor the technical performance of consent mechanisms.
● User Story: As a DF Admin (David) or a power user in Marketing (Mohan), I need robust tools to check my data usage against consent records before I act, and a simple way to manage our relationships with the vendors who process our data.
● Functional Requirements: 1. Consent Validation Console (Manual/Batch V1.0):
○ This feature is the primary tool for business teams to ensure their data processing activities are aligned with the consent provided by Data Principals. It is a preventative control to mitigate the risk of non-compliant processing.
○ Bulk Validation Workfl ow:
1. The UI shall present a clean interface with two main input areas: a fi eld for uploading a CSV fi le and a dropdown menu.
2. Input Method: The user can either:
■ Upload a single-column CSV fi le containing a list of Data Principal identifi ers (email addresses or mobile numbers). The system must validate that the fi le is correctly formatted.
■ Paste a plain-text list of identifi ers directly into a text area.
3. Purpose Selection: The user must select the specifi c processing Purpose ID from a dropdown menu. This menu is dynamically populated from the
active purposes in the Purpose Library (managed by the DPO). This is a mandatory fi eld.
4. Initiation: A "Run Validation" button initiates the background job. The system will provide feedback that the job has started.
○ Validation Results:
1. Upon completion (typically within seconds for reasonably sized lists), the UI will display a clear summary report:
■ Total Records Submitted: [X]
■ Valid for Purpose [P00X - Purpose Name]: [Y]
■ Invalid (Consent Denied or Withdrawn): [Z]
■ Identifi ers Not Found in System: [A]
2. A "Download Valid List" button shall be prominently displayed. Clicking it will download a new CSV fi le containing only the [Y] identifi ers that were confi rmed to have valid, active consent for the selected purpose.
○ Validation History:
1. A separate tab or section will display a log table of all past validation jobs.
2. The table will include Job ID, Timestamp, Initiated By (User ID), Purpose Validated, Total Records Submitted, and a summary of the results (Valid/Invalid). This provides an audit trail for the DPO to review processing activities.
● 2. Processor & Third-Party Management:
○ This feature provides a centralized registry of all external entities that process personal data on behalf of the Data Fiduciary. This registry is the source of truth for all third-party compliance actions.
○ Processor Registry UI:
1. The main view will be a table listing all registered processors with columns for Processor Name, Primary Contact Email, Status (Active/Inactive), and DPA on File (Yes/No).
○ Add/Edit Processor Form:
1. A form will allow an admin to add or edit a processor's details. The fi elds shall be:
■ Processor Name: The legal name of the entity (e.g., "Amazon Web Services EMEA SARL").
■ Primary Contact Email: The offi cial email address for compliance and DPR notifi cations. The system must validate this is a valid email format.
■ Description of Services: A text fi eld to describe the services provided (e.g., "Cloud infrastructure hosting," "External Legal Counsel").
■ DPA Upload: An upload fi eld allowing the user to attach the signed Data Processing Agreement (DPA) or contract in PDF format. The UI will indicate if a fi le is currently associated with the record.
■ Data Categories Processed: A multi-select checklist populated from the Data Lens tags (e.g., "Contact Info," "Financial Info"). This links the processor to the type of data they handle.
● 3. Automated Processor Communication:
○ This is a critical backend workfl ow that automates part of the DPR fulfi llment process, requiring no direct UI interaction during its operation.
○ System Logic:
1. When a Data Principal submits a DPR for Correction or Erasure, the system's workfl ow engine is triggered.
2. The engine queries the Ark Data Lens data map to identify all data fl ows associated with that Data Principal.
3. The system checks if any of these fl ows terminate at a node identifi ed as a Processor/Third-Party in the lineage map.
4. For every match found, the system retrieves the Primary Contact Email from the Processor Registry for that specifi c processor.
5. An email is automatically dispatched to that contact address.
○ Email Template and Content:
1. The email's content will be generated from a master template managed in the System Administration > Email Templates module.
2. The email must be clear and actionable, containing variables that are auto-populated by the system:
■ Subject: "Urgent Data Subject Request from [Data Fiduciary Name] - Ref: [Parent Ticket ID]"
■ Body: Must clearly state:
■ The Data Fiduciary has received a legally binding request from a Data Principal.
■ The Request Type: (Correction or Erasure).
■ The Data Principal Identifi er: (e.g., user@email.com).
■ A clear instruction to take the corresponding action on all personal data associated with this identifi er within their systems.
■ A legal disclaimer regarding their obligations under their DPA with the Data Fiduciary.
3. This automated communication is logged as an event in the history of the Parent Ticket in the DPO Workbench.
4.3.5. Module: Breach Notifi cation
● Description: This module provides the Data Protection Offi cer with a dedicated, end-to-end toolkit for managing the communication and documentation aspects of a personal data breach. In a high-stress scenario, this module enforces a methodical process, ensuring that the Data Fiduciary can meet its time-sensitive notifi cation obligations under DPDPA Section 8(6) and Rule 7 in a manner that is both compliant and auditable.
● User Story: As a DPO (Priya), when a data breach is confi rmed, I need a reliable system to help me defi ne the scope, draft a compliant notice, notify all affected individuals without delay, and generate the necessary documentation for regulatory reporting.
● Functional Requirements: 1. Incident Management Dashboard:
○ This is the central landing page for all breach-related activities, providing the DPO with a high-level overview of all recorded incidents.
○ UI & Functionality:
■ The dashboard shall feature a table listing all recorded data breach incidents.
■ The table must include the following sortable columns: Incident ID (system-generated), Incident Name (user-defi ned), Date Discovered, Status, and Affected Users Count.
■ The Status fi eld will be a system-managed state, such as: "Under Investigation," "Notifi cations Sent," or "Closed."
■ A prominent "Create New Breach Incident" button shall be the primary call to action on this page.
● 2. Incident Creation Wizard:
○ This is a mandatory, step-by-step workfl ow designed to guide the DPO through the formal process of documenting a new breach, ensuring no critical information is missed at the outset.
○ Wizard Steps:
■ Step 1: Incident Details: The DPO must complete the following form fi elds:
■ Incident Name: A mandatory, internal-facing name for easy reference (e.g., "Q3-2025 Customer DB Server Compromise").
■ Date & Time of Discovery: A mandatory date-time picker to accurately record when the organization became aware of the breach.
■ Description of Breach: A mandatory, detailed text area for the DPO to describe the nature of the breach (e.g., "Unauthorized access to production server"), its known or suspected cause, and its current known extent.
■ Step 2: Defi ne Scope of Affected Users: This step is critical for identifying the notifi cation audience. The DPO must choose one of the following methods:
■ Option A: Select Data Source: The UI will present a dropdown menu of all connected Data Sources managed in Ark Data Lens. Selecting a source (e.g., "Production Customer DB") will automatically scope the notifi cation to all Data Principals whose data is mapped to that source. The system will display the total count of affected users found.
■ Option B: Upload a List: If the breach is limited to a specifi c subset of users, the DPO can upload a single-column CSV fi le of their unique identifi ers (emails or mobile numbers). The system will validate the list and display the count of recognized users.
■ Step 3: Review & Create: A fi nal summary screen will display all the information entered. The DPO must confi rm the details before clicking "Create Incident." This action formally records the incident in the system with a unique Incident ID and sets its status to "Under Investigation."
● 3. Notifi cation Drafting & Dispatch:
○ Once an incident is created, the DPO is taken to its dedicated page, where the notifi cation workfl ow begins.
○ Notifi cation Editor:
■ The UI will feature a rich-text editor pre-populated with a template that is compliant with DPDPA Rule 7.
■ The template will contain clearly marked, mandatory sections for the DPO to complete:
■ A description of the breach (nature, extent, timing).
■ The likely consequences relevant to the Data Principal.
■ Measures already implemented or proposed to mitigate risk.
■ Safety measures the Data Principal can take (e.g., "We advise you to monitor your account for suspicious activity").
■ Business contact information (pulled from the DPO's profi le) for further queries.
■ The DPO can save the draft at any time.
○ Dispatch Control:
■ A "Send Notifi cations" button is available once the draft is saved.
■ Clicking this button will trigger a fi nal, high-stakes confi rmation modal: "You are about to send this breach notifi cation to [X] users. This action is irreversible and will be permanently logged. Are you sure you want to proceed?"
■ Upon typing "CONFIRM" into a text box and clicking the fi nal confi rmation, the system queues the notifi cation for bulk dispatch via the confi gured SMTP server.
■ The incident Status automatically updates to "Notifi cations Sent," and the timestamp is logged.
● 4. Regulatory Reporting:
○ This feature is designed to simplify the DPO's obligation to report the breach to the Data Protection Board.
○ Functionality:
■ Within the page for any created incident, a "Generate Incident Report" button will be available.
■ This function will generate a comprehensive, presentation-ready PDF document.
■ The PDF report will be professionally formatted and include:
■ A header with the Data Fiduciary's name and the Incident ID.
■ All details captured during the Incident Creation Wizard.
■ The fi nal, exact text of the notifi cation that was sent to Data Principals.
■ A summary section detailing the Total Users Notifi ed and the Date and Time of Notifi cation.
■ A timeline of key events logged within ComplyArk (e.g., Incident Created, Notifi cations Sent).
■ This downloadable PDF serves as a critical piece of evidence for the DPO's regulatory fi lings and internal post-mortem reviews.
4.3.6. Module: Verifi able Parental Consent
● Description: This module provides the Data Protection Offi cer with a specialized and secure toolkit to manage the high-risk and legally sensitive area of processing personal data for children and persons with disabilities. It is meticulously designed to ensure strict adherence to DPDPA Section 9 by providing a structured, auditable workfl ow for obtaining and verifying consent from a parent or lawful guardian before any processing of the minor's data occurs.
● User Story: As a DPO (Priya), when our service receives a signup request from a minor, I need a dedicated, secure queue where I can review the guardian's provided consent and identity verifi cation, and then formally approve or reject it, with my entire decision process being logged.
● Functional Requirements: 1. The Parental Consent Queue:
○ This feature will appear as a dedicated section in the DPO Workbench, labeled "Parental Consent Approvals."
○ UI & Functionality:
1. The primary interface shall be a table that centralizes all consent requests awaiting guardian verifi cation.
2. The table must contain the following sortable and fi lterable columns: Request ID (system-generated), Child's Identifi er (e.g., child@email.com), Guardian's Identifi er (e.g., parent@email.com), Date Submitted, and Verifi cation Method.
3. The Verifi cation Method column will display one of the system-supported methods used during the request submission, such as: "Existing User," "Manual Upload," or (in the future) "Digilocker."
4. The DPO must be able to fi lter the queue by Verifi cation Method and Request Status (Pending Review, Approved, Rejected) to effi ciently manage their workload.
● 2. The Consent Review Interface:
○ This is the detailed view where the DPO performs the actual verifi cation. It is accessed by clicking on a Request ID from the queue.
○ UI Layout: The interface must be cleanly structured to present all relevant information without clutter.
1. Child's Information Section: Displays the details submitted for the minor, including Name, Email, and Age.
2. Guardian's Information Section: Displays the details submitted for the parent or lawful guardian, including Name and Contact Information.
3. Request Details Section: Shows metadata about the request itself, including Date & Time Submitted and the Verifi cation Method used.
4. Evidence Viewer: This is the most critical component of the review interface. Its content is conditional based on the verifi cation method:
■ If Verifi cation Method is "Manual Upload": This section will contain a secure, in-browser document viewer. It will display the ID proof document(s) (e.g., PDF, JPG) uploaded by the guardian. The viewer must support essential tools like zoom and pan for clear inspection. A "Download Document" link must also be available for offl ine verifi cation if needed. All access to these documents must be logged.
■ If Verifi cation Method is "Existing User": This section will display a system-generated, non-editable confi rmation message: "Guardian's identity was programmatically verifi ed as an existing adult user on [Timestamp]."
■ (Future V1.1) If Verifi cation Method is "Digilocker": This section will display a confi rmation message: "Guardian's identity was successfully verifi ed via Digilocker on [Timestamp]."
● 3. The Decision Workfl ow:
○ This workfl ow provides the DPO with the fi nal, auditable actions to complete the verifi cation process.
○ DPO Actions: At the bottom of the Consent Review Interface, two prominent, mutually exclusive buttons will be displayed: Approve and Reject.
○ Rejection Workfl ow:
1. Clicking the Reject button must trigger a mandatory modal dialog to prevent accidental rejections.
2. The dialog will require the DPO to select a Reason for Rejection from a pre-defi ned, confi gurable dropdown list. Options must include: "ID Proof
Unclear or Invalid," "Guardian Identity Not Verifi able," "Information Mismatch," and "Other."
3. If "Other" is selected, a free-text comment fi eld becomes mandatory.
4. Submitting the rejection form updates the request status to "Rejected," logs the action and the reason in the audit log, and can trigger a templated notifi cation email to the guardian.
○ Approval Workfl ow:
1. Clicking the Approve button will trigger a fi nal confi rmation prompt: "You are about to approve this consent request, enabling the processing of a minor's data. This action is a signifi cant compliance event and will be logged. Are you sure you want to proceed?"
2. Upon confi rmation, the system updates the request status to "Approved."
3. This approval must be recorded as a high-priority event in the main audit log, capturing the Request ID, the DPO's User ID, and the Timestamp.
○ System Integration Logic (Backend):
1. The fi nal "Approved" or "Rejected" status of a consent request must be accessible to the Data Fiduciary's other systems.
2. ComplyArk will expose a secure, internal API endpoint (e.g., GET /api/internal/parental-consent/{Request ID}/status) that the client's own application can query.
3. This allows the client's application to programmatically check if consent has been verifi ed before proceeding with the creation of the child's account or enabling services for them. The responsibility of ComplyArk is to provide the verifi able status; the client's system is responsible for acting on it.
4.3.7. Module: System Administration & Confi guration
● Description: This module is the "engine room" of Ark CMS, accessible only to users with the Data Fiduciary Administrator role. It provides the essential, centralized controls for managing internal users, confi guring system-wide technical behaviors, and customizing automated communications. Effective management of this module is critical for the security, stability, and operational effi ciency of the entire compliance suite.
● User Story: As a DF Admin (David), I need a single, secure area where I can manage my team's access to ComplyArk, set up the email server, and defi ne system-wide rules like password policies and SLA timers, without needing to edit confi guration fi les manually.
● Functional Requirements: 1. Internal User Management:
○ This feature allows the DF Admin to control who has access to the ComplyArk platform and at what level of privilege.
○ User Management UI:
■ The main view shall be a table listing all internal users with the following sortable columns: User Name, Email Address, Assigned Role, and Status.
■ The Status column will display a clear visual indicator for "Active" or "Disabled."
○ User Creation Workfl ow:
■ A "Create New User" button will open a modal form requiring:
■ Full Name (text fi eld).
■ Business Email Address (text fi eld, must be a valid email format). The system must ensure this email is unique across all users.
■ Assign Role: A mandatory dropdown menu containing the pre-defi ned roles (Data Protection Offi cer, Data Fiduciary Administrator, Department User). If Department User is selected, an additional mandatory dropdown will appear to assign them to a specifi c department (departments are confi gured in Data Lens).
■ Upon creation, the system will send a welcome email to the new user's address. This email, generated from a customizable template, will contain a temporary password and a unique, single-use link for their fi rst login, where they will be forced to set a permanent password.
○ User Lifecycle Management:
■ Edit: An "Edit" action next to each user will allow the admin to change their Full Name or Assigned Role.
■ Disable/Enable: A toggle switch will allow the admin to instantly deactivate a user's account, preventing them from logging in. This action is preferable to deletion as it preserves their activity history in the audit logs.
■ Password Reset: A "Reset Password" action will trigger a password reset email to the user, following the same secure process as the initial welcome email.
● 2. Global Confi guration:
○ This feature provides a centralized, tabbed interface for managing all system-wide parameters.
○ UI Layout & Tabs:
■ Tab 1: Infrastructure Confi guration:
■ SMTP Settings: A form with fi elds for SMTP Host, Port, Encryption (None, SSL/TLS), Username, Password (masked fi eld), and a Default "From" Address. A "Send Test Email" button must be present to verify the settings by sending an email to the logged-in admin's address.
■ OTP Settings: A form to confi gure the One-Time Password mechanism, with fi elds for OTP Length (numeric, 4-8), OTP Expiry Time (in minutes), and Maximum Failed OTP Attempts before a temporary lockout.
■ Tab 2: Security Settings:
■ A form with numeric input fi elds to set global security policies: User Session Inactivity Timeout (in minutes), Password Expiry Period (in days) (0 for never), and Maximum Failed Login Attempts before an account is locked.
■ Tab 3: Workfl ow Settings:
■ A form with numeric input fi elds to defi ne the operational timelines for compliance workfl ows: DPR Resolution SLA (in hours) and Grievance Auto-Escalation Period (in hours). These values will power the timers and alerts in the DPO Workbench.
● 3. Email Template Editor:
○ This feature gives the DF Admin the power to customize all automated communications sent by the ComplyArk system, ensuring they align with the company's tone and branding.
○ UI & Functionality:
■ The interface will feature a dropdown menu to select the specifi c email template to be edited (e.g., "New DPR Notifi cation to Department," "OTP Verifi cation Code," "Request Closure Confi rmation to Data Principal").
■ The editor will provide two synchronized text areas:
■ HTML Editor: A rich-text (WYSIWYG) editor for creating visually appealing emails.
■ Plain-Text Editor: A simple text area for the fallback version for email clients that do not render HTML.
■ A visible, context-aware list of available variables/placeholders will be displayed for each template (e.g., for the "Request Closure" template, variables like {user_name}, {request_id}, and {resolution_date} will be shown).
■ A "Send Test Preview" button will allow the admin to send the currently rendered template to their own email address for review before saving. A "Revert to Default" button will discard all changes and restore the original system template.
4.3.8. Module: Immutable Audit & Logging Engine
● Description: This is not just a feature but a foundational guarantee of the ComplyArk suite. It is the system's "unquestionable memory," ensuring that every signifi cant compliance-related action is recorded in a secure, tamper-evident, and chronologically sound manner. The integrity of this log is the ultimate proof of an organization's accountability under the DPDPA.
● User Story: As a DPO (Priya), during a regulatory audit, I need to be able to instantly retrieve a complete, verifi able history of every action taken regarding a specifi c data request, and I must be able_ to prove that this history has not been altered.
● Functional Requirements: 1. The Immutable Audit Log (Backend Architecture):
○ The logging system must be architected as a secure, append-only data structure. The application logic must strictly enforce that log entries can only be created, never edited or deleted.
○ Cryptographic Chaining: This is a non-negotiable requirement. Each new log entry (Log N) must contain a cryptographic hash (e.g., SHA-256) of the complete contents of the preceding log entry (Log N-1). This creates a "blockchain-like" chain where modifying any single past entry would break the entire chain from that point forward.
○ Comprehensive Event Logging: The system must automatically generate a log entry for every compliance-critical event, including but not limited to:
■ ConsentGranted, ConsentWithdrawn, ConsentModifi ed
■ DPRSubmitted, DPRStatusChanged, DPRResolved
■ GrievanceSubmitted, GrievanceStatusChanged
■ NoticeCreated, NoticeVersionFinalized, NoticePublished
■ UserLoginSuccess, UserLoginFailed
■ AdminConfi gChanged (e.g., "SLA updated from 72 to 48 hours")
■ BreachIncidentCreated, BreachNotifi cationSent
■ ParentalConsentApproved, ParentalConsentRejected
○ Log Entry Structure: Each log entry must be a structured record containing at a minimum: Log ID (sequential), Timestamp (UTC, with millisecond precision), Actor (the User ID or "System" that performed the action), Action Type (e.g., "ConsentWithdrawn"), Resource ID (e.g., the specifi c Ticket ID or Notice ID affected), Source IP Address, a Description of the event, and the Previous Log Hash.
● 2. Audit Log Viewer (DPO/DF Admin UI):
○ This UI provides a human-readable interface to the audit log.
○ UI & Functionality:
■ The primary view will be a clean, paginated table displaying the audit log in reverse chronological order.
■ The UI must include advanced fi ltering controls, allowing the user to precisely query the log by a Timestamp Range, Actor, Action Type, or a specifi c Resource ID.
■ An "Export to CSV" function must be available to export the currently fi ltered view for offl ine analysis or submission to auditors.
○ Integrity Verifi cation:
■ The viewer must feature a prominent "Verify Log Integrity" button.
■ When clicked, the system will perform a background job that iterates through the entire log chain from the fi rst to the last entry, re-calculating the hash at each step and comparing it to the stored hash in the next entry.
■ The UI must display a clear result: either a green "Success: Log Integrity Verifi ed" message or a high-alert red "CRITICAL ERROR: Log chain has been tampered with. Discrepancy found at Log ID [X]" message. This action of verifi cation itself will be logged.
● 3. Exception Log Viewer (DF Admin UI):
○ This is a separate, technical logging tool intended purely for troubleshooting application errors. It must be distinct from the compliance audit log.
○ UI & Functionality:
■ A table view will list application-level errors with columns for Timestamp, Error Level (Warning, Error, Critical), Error Message, and the Service/Module where the error occurred.
■ Each row will be expandable to show the full technical stack trace of the error, providing crucial diagnostic information for the SuperAdmin or DF Admin.
● 4. Contextual Audit History Export
○ Description: To support targeted investigations and audits, users need the ability to export the complete history of a single compliance artifact, rather than searching through the entire system log.
○ Functional Requirement: On the detailed view page for any key resource (e.g., a specifi c DPR Ticket, a Notice, a Data Principal profi le), a dedicated "Export Audit History" button shall be present. Clicking this button will generate a fi ltered PDF or CSV report containing every log entry from the master audit log that is associated with that specifi c Resource ID. The report will be chronologically ordered and formatted for easy readability, providing a complete and defensible history of that single item.