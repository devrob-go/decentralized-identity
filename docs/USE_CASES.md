# Use Cases & Implementation Guide

## Overview

The Decentralized Identity & Authentication System enables a wide range of applications across various industries. This document outlines practical use cases, implementation patterns, and real-world scenarios.

---

## 🌐 Web3 Identity for Traditional Applications

### Use Case: Social Media Platform with Web3 Features

**Scenario:** A traditional social media platform wants to add Web3 capabilities while maintaining familiar user experience.

**Implementation:**
```javascript
// Traditional registration flow with DID integration
async function registerUser(userData) {
  // 1. Register user traditionally
  const authResponse = await fetch('/v1/auth/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(userData)
  });
  
  const user = await authResponse.json();
  
  // 2. User automatically gets DID
  console.log('User DID:', user.user.did);
  
  // 3. Can now use both traditional JWT and DID for authentication
  return user;
}

// Later: Use DID for Web3 features
async function verifyUserForNFTMint(userDID, userHash) {
  const verification = await fetch('/api/v1/did/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      did: userDID,
      user_hash: userHash
    })
  });
  
  if (verification.is_valid) {
    // Proceed with NFT minting
    await mintUserNFT(userDID);
  }
}
```

**Benefits:**
- Seamless user experience (no wallet required initially)
- Gradual Web3 adoption
- Verifiable identity for premium features
- Cross-platform identity portability

---

## 🏥 Healthcare Records Management

### Use Case: Patient-Controlled Medical Records

**Scenario:** Patients control their medical records with verifiable access for healthcare providers.

**Implementation:**

```go
// Healthcare provider service
type HealthcareProvider struct {
    didClient *clients.DIDClient
    records   *RecordManager
}

func (h *HealthcareProvider) VerifyPatientAccess(patientDID, recordID string) error {
    // 1. Verify patient DID is valid
    verification, err := h.didClient.VerifyDID(patientDID, "")
    if err != nil {
        return fmt.Errorf("invalid patient DID: %w", err)
    }
    
    if !verification.IsValid {
        return errors.New("patient DID not verified")
    }
    
    // 2. Check patient has authorized access
    if !h.records.HasPatientAuthorized(recordID, patientDID) {
        return errors.New("patient has not authorized access")
    }
    
    return nil
}

func (h *HealthcareProvider) AccessMedicalRecord(patientDID, recordID string) (*MedicalRecord, error) {
    if err := h.VerifyPatientAccess(patientDID, recordID); err != nil {
        return nil, err
    }
    
    // Log access on blockchain for audit trail
    auditEntry := AuditEntry{
        PatientDID:   patientDID,
        RecordID:     recordID,
        AccessedBy:   h.GetProviderDID(),
        Timestamp:    time.Now(),
        ActionType:   "READ",
    }
    
    h.LogToBlockchain(auditEntry)
    
    return h.records.Get(recordID)
}
```

**Smart Contract for Audit Trail:**
```solidity
contract HealthcareAudit {
    struct AccessLog {
        string patientDID;
        string providerDID;
        string recordID;
        uint256 timestamp;
        string actionType;
    }
    
    mapping(string => AccessLog[]) public patientAccessLogs;
    
    function logAccess(
        string memory patientDID,
        string memory providerDID,
        string memory recordID,
        string memory actionType
    ) external {
        AccessLog memory log = AccessLog({
            patientDID: patientDID,
            providerDID: providerDID,
            recordID: recordID,
            timestamp: block.timestamp,
            actionType: actionType
        });
        
        patientAccessLogs[patientDID].push(log);
    }
}
```

**Benefits:**
- Patient-controlled access
- Immutable audit trail
- HIPAA compliance through cryptographic verification
- Cross-provider identity verification

---

## 🏦 Financial Services & KYC

### Use Case: Cross-Border Banking with Verified Identity

**Scenario:** Banks need to verify customer identity across jurisdictions for regulatory compliance.

**Implementation:**

```python
import hashlib
import json
from datetime import datetime

class KYCVerificationService:
    def __init__(self, did_client):
        self.did_client = did_client
        
    def verify_customer_identity(self, customer_data, customer_did):
        """Verify customer identity using DID"""
        
        # 1. Generate hash from customer data
        customer_hash = self.generate_customer_hash(customer_data)
        
        # 2. Verify DID matches customer data
        verification = self.did_client.verify_did(customer_did, customer_hash)
        
        if not verification['data']['is_valid']:
            raise ValueError("Customer DID verification failed")
            
        # 3. Check DID status on blockchain
        status = self.did_client.get_did_status(customer_did)
        
        if status['data']['status'] != 'registered':
            raise ValueError("Customer DID not registered on blockchain")
            
        # 4. Generate KYC compliance record
        kyc_record = {
            'customer_did': customer_did,
            'verification_time': datetime.utcnow().isoformat(),
            'verification_status': 'VERIFIED',
            'blockchain_tx': status['data']['blockchain_tx'],
            'compliance_level': 'FULL_KYC'
        }
        
        return kyc_record
        
    def generate_customer_hash(self, customer_data):
        """Generate consistent hash from customer data"""
        normalized_data = {
            'name': customer_data['name'].strip().lower(),
            'email': customer_data['email'].strip().lower(),
            'date_of_birth': customer_data['date_of_birth'],
            'document_number': customer_data['document_number']
        }
        
        data_string = json.dumps(normalized_data, sort_keys=True)
        return hashlib.sha256(data_string.encode()).hexdigest()

# Usage in banking application
def onboard_customer(customer_data):
    kyc_service = KYCVerificationService(did_client)
    
    try:
        # Verify customer identity using DID
        kyc_record = kyc_service.verify_customer_identity(
            customer_data, 
            customer_data['did']
        )
        
        # Store KYC record for compliance
        compliance_db.store_kyc_record(kyc_record)
        
        # Approve account opening
        return approve_account_opening(customer_data, kyc_record)
        
    except ValueError as e:
        # Reject application due to identity verification failure
        return reject_application(str(e))
```

**Benefits:**
- Regulatory compliance automation
- Cross-border identity verification
- Immutable compliance records
- Reduced KYC processing time

---

## 🎓 Educational Credentials

### Use Case: Verifiable Academic Credentials

**Scenario:** Universities issue tamper-proof digital diplomas that can be verified globally.

**Implementation:**

```go
package education

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

type AcademicCredential struct {
    StudentDID    string    `json:"student_did"`
    InstitutionDID string   `json:"institution_did"`
    Degree        string    `json:"degree"`
    Major         string    `json:"major"`
    GraduationDate time.Time `json:"graduation_date"`
    CredentialHash string   `json:"credential_hash"`
    BlockchainTx  string    `json:"blockchain_tx"`
}

type EducationService struct {
    didClient *DIDClient
    blockchain *BlockchainClient
}

func (e *EducationService) IssueCredential(studentDID, degree, major string) (*AcademicCredential, error) {
    // 1. Verify student DID
    verification, err := e.didClient.VerifyDID(studentDID, "")
    if err != nil {
        return nil, fmt.Errorf("failed to verify student DID: %w", err)
    }
    
    if !verification.IsValid {
        return nil, fmt.Errorf("invalid student DID")
    }
    
    // 2. Create credential
    credential := &AcademicCredential{
        StudentDID:     studentDID,
        InstitutionDID: e.GetInstitutionDID(),
        Degree:         degree,
        Major:          major,
        GraduationDate: time.Now(),
    }
    
    // 3. Generate credential hash
    credential.CredentialHash = e.generateCredentialHash(credential)
    
    // 4. Store on blockchain
    txHash, err := e.blockchain.StoreCredential(credential)
    if err != nil {
        return nil, fmt.Errorf("failed to store credential on blockchain: %w", err)
    }
    
    credential.BlockchainTx = txHash
    return credential, nil
}

func (e *EducationService) VerifyCredential(credentialHash string) (*AcademicCredential, error) {
    // Retrieve and verify from blockchain
    credential, err := e.blockchain.GetCredential(credentialHash)
    if err != nil {
        return nil, fmt.Errorf("credential not found on blockchain: %w", err)
    }
    
    // Verify credential hash
    expectedHash := e.generateCredentialHash(credential)
    if expectedHash != credential.CredentialHash {
        return nil, fmt.Errorf("credential hash mismatch - tampered credential")
    }
    
    // Verify student DID is still valid
    verification, err := e.didClient.VerifyDID(credential.StudentDID, "")
    if err != nil {
        return nil, fmt.Errorf("failed to verify student DID: %w", err)
    }
    
    if !verification.IsValid {
        return nil, fmt.Errorf("student DID is no longer valid")
    }
    
    return credential, nil
}

func (e *EducationService) generateCredentialHash(cred *AcademicCredential) string {
    data := fmt.Sprintf("%s:%s:%s:%s:%d",
        cred.StudentDID,
        cred.InstitutionDID,
        cred.Degree,
        cred.Major,
        cred.GraduationDate.Unix(),
    )
    
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

**Smart Contract for Credential Storage:**
```solidity
contract EducationCredentials {
    struct Credential {
        string studentDID;
        string institutionDID;
        string credentialHash;
        uint256 issueDate;
        bool isValid;
    }
    
    mapping(string => Credential) public credentials;
    mapping(string => bool) public authorizedInstitutions;
    
    modifier onlyAuthorizedInstitution() {
        require(authorizedInstitutions[msg.sender], "Not authorized institution");
        _;
    }
    
    function issueCredential(
        string memory credentialHash,
        string memory studentDID,
        string memory institutionDID
    ) external onlyAuthorizedInstitution {
        credentials[credentialHash] = Credential({
            studentDID: studentDID,
            institutionDID: institutionDID,
            credentialHash: credentialHash,
            issueDate: block.timestamp,
            isValid: true
        });
    }
    
    function verifyCredential(string memory credentialHash) 
        external view returns (bool, string memory, uint256) {
        Credential memory cred = credentials[credentialHash];
        return (cred.isValid, cred.studentDID, cred.issueDate);
    }
}
```

**Benefits:**
- Tamper-proof credentials
- Global verification without contacting issuing institution
- Reduced credential fraud
- Automated verification for employers

---

## 🏭 Supply Chain Traceability

### Use Case: Food Safety and Origin Verification

**Scenario:** Track food products from farm to table with verifiable provenance.

**Implementation:**

```typescript
interface SupplyChainEvent {
  eventId: string;
  productDID: string;
  participantDID: string;
  eventType: 'HARVEST' | 'PROCESS' | 'TRANSPORT' | 'RETAIL' | 'PURCHASE';
  location: string;
  timestamp: Date;
  metadata: Record<string, any>;
  signature: string;
}

class SupplyChainTracker {
  private didClient: DIDClient;
  private blockchain: BlockchainClient;
  
  constructor(didClient: DIDClient, blockchain: BlockchainClient) {
    this.didClient = didClient;
    this.blockchain = blockchain;
  }
  
  async createProductDID(productData: {
    name: string;
    origin: string;
    producer: string;
  }): Promise<string> {
    // Create unique DID for product
    const productHash = this.generateProductHash(productData);
    const productDID = `did:supply:product:${productHash}`;
    
    // Register on blockchain
    await this.blockchain.registerProductDID(productDID, productData);
    
    return productDID;
  }
  
  async recordSupplyChainEvent(event: SupplyChainEvent): Promise<void> {
    // 1. Verify participant DID
    const verification = await this.didClient.verifyDID(event.participantDID, "");
    if (!verification.data.is_valid) {
      throw new Error(`Invalid participant DID: ${event.participantDID}`);
    }
    
    // 2. Verify product DID exists
    const productExists = await this.blockchain.productExists(event.productDID);
    if (!productExists) {
      throw new Error(`Product DID not found: ${event.productDID}`);
    }
    
    // 3. Generate event signature
    event.signature = await this.signEvent(event);
    
    // 4. Store event on blockchain
    await this.blockchain.recordEvent(event);
    
    console.log(`Supply chain event recorded: ${event.eventType} for ${event.productDID}`);
  }
  
  async traceProduct(productDID: string): Promise<SupplyChainEvent[]> {
    // Get all events for product from blockchain
    const events = await this.blockchain.getProductEvents(productDID);
    
    // Verify each event signature
    for (const event of events) {
      const isValid = await this.verifyEventSignature(event);
      if (!isValid) {
        console.warn(`Invalid signature for event ${event.eventId}`);
      }
    }
    
    return events.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime());
  }
  
  async verifyProductOrigin(productDID: string, claimedOrigin: string): Promise<boolean> {
    const trace = await this.traceProduct(productDID);
    
    // Find harvest event (origin)
    const harvestEvent = trace.find(event => event.eventType === 'HARVEST');
    if (!harvestEvent) {
      return false;
    }
    
    return harvestEvent.location === claimedOrigin;
  }
  
  private generateProductHash(productData: any): string {
    const dataString = JSON.stringify(productData, Object.keys(productData).sort());
    return crypto.createHash('sha256').update(dataString).digest('hex').substring(0, 16);
  }
  
  private async signEvent(event: SupplyChainEvent): Promise<string> {
    const eventData = `${event.productDID}:${event.participantDID}:${event.eventType}:${event.timestamp.toISOString()}`;
    return crypto.createHash('sha256').update(eventData).digest('hex');
  }
  
  private async verifyEventSignature(event: SupplyChainEvent): Promise<boolean> {
    const expectedSignature = await this.signEvent(event);
    return event.signature === expectedSignature;
  }
}

// Usage example
async function trackFarmToTable() {
  const tracker = new SupplyChainTracker(didClient, blockchain);
  
  // 1. Create product DID at farm
  const productDID = await tracker.createProductDID({
    name: "Organic Tomatoes",
    origin: "Green Valley Farm, CA",
    producer: "did:farmer:john_doe"
  });
  
  // 2. Record harvest
  await tracker.recordSupplyChainEvent({
    eventId: "harvest_001",
    productDID: productDID,
    participantDID: "did:farmer:john_doe",
    eventType: "HARVEST",
    location: "Green Valley Farm, CA",
    timestamp: new Date(),
    metadata: {
      quantity: "100 lbs",
      quality_grade: "A",
      harvest_method: "hand_picked"
    },
    signature: ""
  });
  
  // 3. Record processing
  await tracker.recordSupplyChainEvent({
    eventId: "process_001",
    productDID: productDID,
    participantDID: "did:processor:fresh_foods",
    eventType: "PROCESS",
    location: "Fresh Foods Processing, CA",
    timestamp: new Date(),
    metadata: {
      process_type: "washing_packaging",
      batch_number: "FF2025001"
    },
    signature: ""
  });
  
  // 4. Consumer verification
  const isOriginalFarm = await tracker.verifyProductOrigin(productDID, "Green Valley Farm, CA");
  console.log(`Product origin verified: ${isOriginalFarm}`);
  
  // 5. Full trace
  const fullTrace = await tracker.traceProduct(productDID);
  console.log("Complete supply chain trace:", fullTrace);
}
```

**Benefits:**
- End-to-end traceability
- Food safety verification
- Fraud prevention
- Consumer trust through transparency
- Rapid contamination source identification

---

## 🎫 Event Ticketing & Access Control

### Use Case: Concert Tickets with Anti-Fraud Protection

**Scenario:** Event organizers issue verifiable tickets that prevent counterfeiting and enable secure resale.

**Implementation:**

```python
import uuid
import hashlib
import json
from datetime import datetime, timedelta

class EventTicketingSystem:
    def __init__(self, did_client, blockchain_client):
        self.did_client = did_client
        self.blockchain = blockchain_client
        self.events = {}
        self.tickets = {}
        
    def create_event(self, organizer_did, event_data):
        """Create a new event with DID-based access control"""
        
        # Verify organizer DID
        verification = self.did_client.verify_did(organizer_did, "")
        if not verification['data']['is_valid']:
            raise ValueError("Invalid organizer DID")
            
        event_id = str(uuid.uuid4())
        event = {
            'id': event_id,
            'organizer_did': organizer_did,
            'name': event_data['name'],
            'venue': event_data['venue'],
            'date': event_data['date'],
            'max_tickets': event_data['max_tickets'],
            'price': event_data['price'],
            'created_at': datetime.utcnow().isoformat()
        }
        
        # Store event on blockchain for immutability
        tx_hash = self.blockchain.store_event(event)
        event['blockchain_tx'] = tx_hash
        
        self.events[event_id] = event
        return event
        
    def issue_ticket(self, event_id, buyer_did, payment_proof):
        """Issue a ticket to a verified buyer"""
        
        if event_id not in self.events:
            raise ValueError("Event not found")
            
        event = self.events[event_id]
        
        # Verify buyer DID
        verification = self.did_client.verify_did(buyer_did, "")
        if not verification['data']['is_valid']:
            raise ValueError("Invalid buyer DID")
            
        # Check ticket availability
        issued_count = len([t for t in self.tickets.values() if t['event_id'] == event_id])
        if issued_count >= event['max_tickets']:
            raise ValueError("Event sold out")
            
        # Generate unique ticket
        ticket_id = str(uuid.uuid4())
        ticket_data = {
            'id': ticket_id,
            'event_id': event_id,
            'buyer_did': buyer_did,
            'original_buyer_did': buyer_did,
            'issue_date': datetime.utcnow().isoformat(),
            'price_paid': event['price'],
            'payment_proof': payment_proof,
            'status': 'VALID',
            'transfer_count': 0
        }
        
        # Generate ticket hash for verification
        ticket_data['hash'] = self.generate_ticket_hash(ticket_data)
        
        # Store ticket on blockchain
        tx_hash = self.blockchain.store_ticket(ticket_data)
        ticket_data['blockchain_tx'] = tx_hash
        
        self.tickets[ticket_id] = ticket_data
        return ticket_data
        
    def transfer_ticket(self, ticket_id, current_owner_did, new_owner_did, transfer_price):
        """Transfer ticket to new owner with verification"""
        
        if ticket_id not in self.tickets:
            raise ValueError("Ticket not found")
            
        ticket = self.tickets[ticket_id]
        
        # Verify current owner
        if ticket['buyer_did'] != current_owner_did:
            raise ValueError("Not current ticket owner")
            
        # Verify current owner DID
        verification = self.did_client.verify_did(current_owner_did, "")
        if not verification['data']['is_valid']:
            raise ValueError("Invalid current owner DID")
            
        # Verify new owner DID
        verification = self.did_client.verify_did(new_owner_did, "")
        if not verification['data']['is_valid']:
            raise ValueError("Invalid new owner DID")
            
        # Check transfer limits (prevent scalping)
        if ticket['transfer_count'] >= 3:
            raise ValueError("Maximum transfers exceeded")
            
        # Update ticket ownership
        ticket['buyer_did'] = new_owner_did
        ticket['transfer_count'] += 1
        ticket['last_transfer_date'] = datetime.utcnow().isoformat()
        ticket['last_transfer_price'] = transfer_price
        
        # Update hash and blockchain record
        ticket['hash'] = self.generate_ticket_hash(ticket)
        tx_hash = self.blockchain.update_ticket(ticket)
        ticket['blockchain_tx'] = tx_hash
        
        return ticket
        
    def verify_ticket_at_entry(self, ticket_id, presented_did):
        """Verify ticket at event entry"""
        
        if ticket_id not in self.tickets:
            return False, "Ticket not found"
            
        ticket = self.tickets[ticket_id]
        
        # Check ticket status
        if ticket['status'] != 'VALID':
            return False, f"Ticket status: {ticket['status']}"
            
        # Verify presenter owns the ticket
        if ticket['buyer_did'] != presented_did:
            return False, "DID does not match ticket owner"
            
        # Verify presenter's DID
        verification = self.did_client.verify_did(presented_did, "")
        if not verification['data']['is_valid']:
            return False, "Invalid DID"
            
        # Verify ticket hash hasn't been tampered with
        expected_hash = self.generate_ticket_hash(ticket)
        if ticket['hash'] != expected_hash:
            return False, "Ticket has been tampered with"
            
        # Check event hasn't started yet (prevent early entry)
        event = self.events[ticket['event_id']]
        event_date = datetime.fromisoformat(event['date'])
        if datetime.utcnow() < event_date - timedelta(hours=2):
            return False, "Entry not yet permitted"
            
        # Mark ticket as used
        ticket['status'] = 'USED'
        ticket['entry_time'] = datetime.utcnow().isoformat()
        
        # Update blockchain
        tx_hash = self.blockchain.update_ticket(ticket)
        
        return True, "Entry granted"
        
    def generate_ticket_hash(self, ticket_data):
        """Generate hash for ticket verification"""
        hash_data = {
            'id': ticket_data['id'],
            'event_id': ticket_data['event_id'],
            'original_buyer_did': ticket_data['original_buyer_did'],
            'issue_date': ticket_data['issue_date'],
            'transfer_count': ticket_data['transfer_count']
        }
        
        data_string = json.dumps(hash_data, sort_keys=True)
        return hashlib.sha256(data_string.encode()).hexdigest()

# Usage example
def run_concert_ticketing():
    ticketing = EventTicketingSystem(did_client, blockchain_client)
    
    # 1. Create event
    event = ticketing.create_event(
        "did:organizer:live_nation",
        {
            'name': "Rock Concert 2025",
            'venue': "Madison Square Garden",
            'date': "2025-12-31T20:00:00",
            'max_tickets': 20000,
            'price': 150.00
        }
    )
    
    # 2. Issue ticket to fan
    ticket = ticketing.issue_ticket(
        event['id'],
        "did:fan:alice_smith",
        "payment_tx_12345"
    )
    
    # 3. Fan transfers ticket to friend
    transferred_ticket = ticketing.transfer_ticket(
        ticket['id'],
        "did:fan:alice_smith",
        "did:fan:bob_jones",
        175.00
    )
    
    # 4. Verify at entry
    is_valid, message = ticketing.verify_ticket_at_entry(
        ticket['id'],
        "did:fan:bob_jones"
    )
    
    print(f"Entry verification: {is_valid} - {message}")
```

**Benefits:**
- Counterfeit-proof tickets
- Transparent secondary market
- Automated entry verification
- Fraud prevention
- Fair pricing enforcement

---

## 💼 Corporate Identity & Employee Verification

### Use Case: Remote Employee Identity Verification

**Scenario:** Companies need to verify remote employee identities for secure access to systems and facilities.

**Implementation:**

```go
package corporate

import (
    "fmt"
    "time"
)

type EmployeeCredential struct {
    EmployeeDID    string    `json:"employee_did"`
    CompanyDID     string    `json:"company_did"`
    EmployeeID     string    `json:"employee_id"`
    Department     string    `json:"department"`
    Role           string    `json:"role"`
    SecurityLevel  string    `json:"security_level"`
    HireDate       time.Time `json:"hire_date"`
    ExpiryDate     time.Time `json:"expiry_date"`
    IsActive       bool      `json:"is_active"`
    BlockchainTx   string    `json:"blockchain_tx"`
}

type CorporateIdentityService struct {
    didClient  *DIDClient
    blockchain *BlockchainClient
    hrSystem   *HRSystem
}

func (c *CorporateIdentityService) OnboardEmployee(employeeData EmployeeData) (*EmployeeCredential, error) {
    // 1. Create DID for employee
    employeeDID, err := c.didClient.CreateDID(&DIDCreateRequest{
        UserID:   employeeData.ID,
        Name:     employeeData.FullName,
        Email:    employeeData.WorkEmail,
        Password: employeeData.TempPassword,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create employee DID: %w", err)
    }
    
    // 2. Create employee credential
    credential := &EmployeeCredential{
        EmployeeDID:   employeeDID.Data.DID.DID,
        CompanyDID:    c.GetCompanyDID(),
        EmployeeID:    employeeData.ID,
        Department:    employeeData.Department,
        Role:          employeeData.Role,
        SecurityLevel: c.determineSecurityLevel(employeeData.Role),
        HireDate:      time.Now(),
        ExpiryDate:    time.Now().AddDate(1, 0, 0), // 1 year
        IsActive:      true,
    }
    
    // 3. Store credential on blockchain
    txHash, err := c.blockchain.StoreEmployeeCredential(credential)
    if err != nil {
        return nil, fmt.Errorf("failed to store credential on blockchain: %w", err)
    }
    
    credential.BlockchainTx = txHash
    
    // 4. Update HR system
    err = c.hrSystem.UpdateEmployeeDID(employeeData.ID, employeeDID.Data.DID.DID)
    if err != nil {
        // Log error but don't fail - credential is still valid
        fmt.Printf("Warning: failed to update HR system: %v\n", err)
    }
    
    return credential, nil
}

func (c *CorporateIdentityService) VerifyEmployeeAccess(employeeDID, resourceID string, requiredLevel string) (bool, error) {
    // 1. Verify employee DID is valid
    verification, err := c.didClient.VerifyDID(employeeDID, "")
    if err != nil {
        return false, fmt.Errorf("failed to verify employee DID: %w", err)
    }
    
    if !verification.Data.IsValid {
        return false, fmt.Errorf("invalid employee DID")
    }
    
    // 2. Get employee credential from blockchain
    credential, err := c.blockchain.GetEmployeeCredential(employeeDID)
    if err != nil {
        return false, fmt.Errorf("failed to get employee credential: %w", err)
    }
    
    // 3. Check if credential is active and not expired
    if !credential.IsActive {
        return false, fmt.Errorf("employee credential is inactive")
    }
    
    if time.Now().After(credential.ExpiryDate) {
        return false, fmt.Errorf("employee credential has expired")
    }
    
    // 4. Check security level
    if !c.hasRequiredSecurityLevel(credential.SecurityLevel, requiredLevel) {
        return false, fmt.Errorf("insufficient security level")
    }
    
    // 5. Log access attempt
    c.logAccessAttempt(employeeDID, resourceID, "GRANTED")
    
    return true, nil
}

func (c *CorporateIdentityService) RevokeEmployeeAccess(employeeDID, reason string) error {
    // Get current credential
    credential, err := c.blockchain.GetEmployeeCredential(employeeDID)
    if err != nil {
        return fmt.Errorf("failed to get employee credential: %w", err)
    }
    
    // Mark as inactive
    credential.IsActive = false
    
    // Update on blockchain
    txHash, err := c.blockchain.UpdateEmployeeCredential(credential)
    if err != nil {
        return fmt.Errorf("failed to revoke credential on blockchain: %w", err)
    }
    
    // Log revocation
    c.logSecurityEvent(employeeDID, "CREDENTIAL_REVOKED", reason, txHash)
    
    return nil
}

func (c *CorporateIdentityService) GenerateAccessReport(departmentID string, fromDate, toDate time.Time) (*AccessReport, error) {
    // Get all employees in department
    employees, err := c.hrSystem.GetEmployeesByDepartment(departmentID)
    if err != nil {
        return nil, fmt.Errorf("failed to get department employees: %w", err)
    }
    
    report := &AccessReport{
        Department: departmentID,
        FromDate:   fromDate,
        ToDate:     toDate,
        Employees:  make([]EmployeeAccessSummary, 0),
    }
    
    for _, employee := range employees {
        // Get access logs for employee
        logs, err := c.getAccessLogs(employee.DID, fromDate, toDate)
        if err != nil {
            continue // Skip on error, log it
        }
        
        summary := EmployeeAccessSummary{
            EmployeeDID:   employee.DID,
            EmployeeName:  employee.Name,
            AccessCount:   len(logs),
            LastAccess:    c.getLastAccessTime(logs),
            SecurityLevel: employee.SecurityLevel,
        }
        
        report.Employees = append(report.Employees, summary)
    }
    
    return report, nil
}

// Smart contract integration
func (c *CorporateIdentityService) VerifyEmployeeOnChain(employeeDID string) (bool, error) {
    // Call smart contract to verify employee
    result, err := c.blockchain.CallContract("verifyEmployee", employeeDID)
    if err != nil {
        return false, err
    }
    
    return result.(bool), nil
}

func (c *CorporateIdentityService) determineSecurityLevel(role string) string {
    switch role {
    case "CEO", "CTO", "CISO":
        return "EXECUTIVE"
    case "Manager", "Director":
        return "MANAGEMENT"
    case "Security", "IT Admin":
        return "HIGH"
    case "Developer", "Analyst":
        return "MEDIUM"
    default:
        return "BASIC"
    }
}

func (c *CorporateIdentityService) hasRequiredSecurityLevel(employeeLevel, requiredLevel string) bool {
    levels := map[string]int{
        "BASIC":      1,
        "MEDIUM":     2,
        "HIGH":       3,
        "MANAGEMENT": 4,
        "EXECUTIVE":  5,
    }
    
    return levels[employeeLevel] >= levels[requiredLevel]
}
```

**Smart Contract for Employee Verification:**
```solidity
contract CorporateIdentity {
    struct Employee {
        string employeeDID;
        string companyDID;
        string role;
        uint256 securityLevel;
        uint256 hireDate;
        uint256 expiryDate;
        bool isActive;
    }
    
    mapping(string => Employee) public employees;
    mapping(string => bool) public authorizedCompanies;
    
    event EmployeeRegistered(string employeeDID, string companyDID, uint256 timestamp);
    event EmployeeRevoked(string employeeDID, string reason, uint256 timestamp);
    
    function registerEmployee(
        string memory employeeDID,
        string memory companyDID,
        string memory role,
        uint256 securityLevel,
        uint256 expiryDate
    ) external {
        require(authorizedCompanies[companyDID], "Company not authorized");
        
        employees[employeeDID] = Employee({
            employeeDID: employeeDID,
            companyDID: companyDID,
            role: role,
            securityLevel: securityLevel,
            hireDate: block.timestamp,
            expiryDate: expiryDate,
            isActive: true
        });
        
        emit EmployeeRegistered(employeeDID, companyDID, block.timestamp);
    }
    
    function verifyEmployee(string memory employeeDID) external view returns (bool) {
        Employee memory emp = employees[employeeDID];
        return emp.isActive && block.timestamp <= emp.expiryDate;
    }
    
    function revokeEmployee(string memory employeeDID, string memory reason) external {
        Employee storage emp = employees[employeeDID];
        require(bytes(emp.employeeDID).length > 0, "Employee not found");
        
        emp.isActive = false;
        emit EmployeeRevoked(employeeDID, reason, block.timestamp);
    }
}
```

**Benefits:**
- Tamper-proof employee credentials
- Automated access control
- Compliance audit trails
- Remote identity verification
- Cross-company employee verification

---

## 🔗 Integration Patterns

### Common Integration Patterns

1. **Hybrid Authentication**: Traditional + DID
2. **Progressive Enhancement**: Add DID to existing systems
3. **DID-First**: New systems built around DIDs
4. **Cross-Chain**: Multiple blockchain integration
5. **Off-Chain Verification**: Local verification with blockchain backup

### Performance Considerations

- **Caching**: Cache DID verification results
- **Batch Processing**: Process multiple DIDs together
- **Async Operations**: Use queues for blockchain operations
- **Fallback**: Local verification when blockchain unavailable

### Security Best Practices

- **Key Management**: Secure private key storage
- **Rotation**: Regular key rotation policies
- **Monitoring**: Real-time fraud detection
- **Compliance**: Regulatory requirement adherence

These use cases demonstrate the flexibility and power of the DID system across various industries and scenarios. Each implementation can be customized based on specific requirements and regulatory needs.
