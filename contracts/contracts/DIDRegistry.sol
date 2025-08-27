// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

/**
 * @title DIDRegistry
 * @dev A smart contract for managing Decentralized Identifiers (DIDs) on Ethereum
 * @dev This contract stores DID registrations and provides verification functionality
 */
contract DIDRegistry is Ownable, ReentrancyGuard {
    using Strings for string;

    // Events
    event DIDRegistered(bytes32 indexed userHash, string did, uint256 timestamp);
    event DIDUpdated(bytes32 indexed userHash, string did, uint256 timestamp);
    event DIDRevoked(bytes32 indexed userHash, string did, uint256 timestamp);
    event AdminChanged(address indexed previousAdmin, address indexed newAdmin);

    // Structs
    struct DIDRecord {
        string did;
        uint256 registrationTime;
        uint256 lastUpdateTime;
        bool isActive;
        bool isRevoked;
        string metadata; // Additional metadata as JSON string
    }

    // State variables
    mapping(bytes32 => DIDRecord) public didRecords;
    mapping(string => bytes32) public didToUserHash;
    mapping(address => bool) public authorizedOperators;
    
    uint256 public totalDIDs;
    uint256 public activeDIDs;
    uint256 public revokedDIDs;

    // Modifiers
    modifier onlyAuthorized() {
        require(
            msg.sender == owner() || authorizedOperators[msg.sender],
            "DIDRegistry: caller is not authorized"
        );
        _;
    }

    modifier didExists(bytes32 userHash) {
        require(
            didRecords[userHash].registrationTime > 0,
            "DIDRegistry: DID does not exist"
        );
        _;
    }

    modifier didNotExist(bytes32 userHash) {
        require(
            didRecords[userHash].registrationTime == 0,
            "DIDRegistry: DID already exists"
        );
        _;
    }

    modifier didNotRevoked(bytes32 userHash) {
        require(
            !didRecords[userHash].isRevoked,
            "DIDRegistry: DID is revoked"
        );
        _;
    }

    // Constructor
    constructor() {
        totalDIDs = 0;
        activeDIDs = 0;
        revokedDIDs = 0;
    }

    /**
     * @dev Register a new DID
     * @param userHash The hash of user identity data
     * @param did The Decentralized Identifier string
     * @param metadata Additional metadata as JSON string
     */
    function registerDID(
        bytes32 userHash,
        string calldata did,
        string calldata metadata
    ) external onlyAuthorized didNotExist(userHash) nonReentrant {
        require(bytes(did).length > 0, "DIDRegistry: DID cannot be empty");
        require(
            didToUserHash[did] == bytes32(0),
            "DIDRegistry: DID already registered"
        );

        DIDRecord memory newRecord = DIDRecord({
            did: did,
            registrationTime: block.timestamp,
            lastUpdateTime: block.timestamp,
            isActive: true,
            isRevoked: false,
            metadata: metadata
        });

        didRecords[userHash] = newRecord;
        didToUserHash[did] = userHash;

        totalDIDs++;
        activeDIDs++;

        emit DIDRegistered(userHash, did, block.timestamp);
    }

    /**
     * @dev Internal function to register a DID without access control
     * @param userHash The hash of user identity data
     * @param did The DID string
     * @param metadata Associated metadata
     */
    function _registerDIDInternal(
        bytes32 userHash,
        string memory did,
        string memory metadata
    ) internal {
        require(bytes(did).length > 0, "DIDRegistry: DID cannot be empty");
        require(
            didToUserHash[did] == bytes32(0),
            "DIDRegistry: DID already registered"
        );
        require(
            didRecords[userHash].registrationTime == 0,
            "DIDRegistry: User hash already has a DID"
        );

        DIDRecord memory newRecord = DIDRecord({
            did: did,
            registrationTime: block.timestamp,
            lastUpdateTime: block.timestamp,
            isActive: true,
            isRevoked: false,
            metadata: metadata
        });

        didRecords[userHash] = newRecord;
        didToUserHash[did] = userHash;

        totalDIDs++;
        activeDIDs++;

        emit DIDRegistered(userHash, did, block.timestamp);
    }

    /**
     * @dev Update an existing DID
     * @param userHash The hash of user identity data
     * @param newDid The new DID string
     * @param metadata Updated metadata
     */
    function updateDID(
        bytes32 userHash,
        string calldata newDid,
        string calldata metadata
    ) external onlyAuthorized didExists(userHash) didNotRevoked(userHash) nonReentrant {
        require(bytes(newDid).length > 0, "DIDRegistry: DID cannot be empty");
        
        string memory oldDid = didRecords[userHash].did;
        
        // Remove old DID mapping if it's different
        if (keccak256(bytes(oldDid)) != keccak256(bytes(newDid))) {
            delete didToUserHash[oldDid];
            didToUserHash[newDid] = userHash;
        }

        // Update the record
        didRecords[userHash].did = newDid;
        didRecords[userHash].lastUpdateTime = block.timestamp;
        didRecords[userHash].metadata = metadata;

        emit DIDUpdated(userHash, newDid, block.timestamp);
    }

    /**
     * @dev Revoke a DID
     * @param userHash The hash of user identity data
     */
    function revokeDID(bytes32 userHash) 
        external 
        onlyAuthorized 
        didExists(userHash) 
        didNotRevoked(userHash) 
        nonReentrant 
    {
        didRecords[userHash].isActive = false;
        didRecords[userHash].isRevoked = true;
        didRecords[userHash].lastUpdateTime = block.timestamp;

        // Remove DID mapping
        delete didToUserHash[didRecords[userHash].did];

        activeDIDs--;
        revokedDIDs++;

        emit DIDRevoked(userHash, didRecords[userHash].did, block.timestamp);
    }

    /**
     * @dev Verify if a DID is valid
     * @param did The DID to verify
     * @return isValid True if the DID is valid and active
     * @return userHash The user hash associated with the DID
     */
    function verifyDID(string calldata did) 
        external 
        view 
        returns (bool isValid, bytes32 userHash) 
    {
        userHash = didToUserHash[did];
        
        if (userHash == bytes32(0)) {
            return (false, bytes32(0));
        }

        DIDRecord memory record = didRecords[userHash];
        isValid = record.isActive && !record.isRevoked;
        
        return (isValid, userHash);
    }

    /**
     * @dev Get DID record by user hash
     * @param userHash The hash of user identity data
     * @return The DID record
     */
    function getDIDRecord(bytes32 userHash) 
        external 
        view 
        didExists(userHash) 
        returns (DIDRecord memory) 
    {
        return didRecords[userHash];
    }

    /**
     * @dev Check if a DID exists
     * @param did The DID to check
     * @return True if the DID exists
     */
    function didExistsPublic(string calldata did) external view returns (bool) {
        return didToUserHash[did] != bytes32(0);
    }

    /**
     * @dev Get user hash by DID
     * @param did The DID
     * @return The user hash
     */
    function getUserHashByDID(string calldata did) external view returns (bytes32) {
        return didToUserHash[did];
    }

    /**
     * @dev Add an authorized operator
     * @param operator The address to authorize
     */
    function addAuthorizedOperator(address operator) external onlyOwner {
        require(operator != address(0), "DIDRegistry: invalid operator address");
        authorizedOperators[operator] = true;
        emit AdminChanged(address(0), operator);
    }

    /**
     * @dev Remove an authorized operator
     * @param operator The address to revoke authorization from
     */
    function removeAuthorizedOperator(address operator) external onlyOwner {
        require(operator != address(0), "DIDRegistry: invalid operator address");
        authorizedOperators[operator] = false;
        emit AdminChanged(operator, address(0));
    }

    /**
     * @dev Get contract statistics
     * @return total Total number of DIDs
     * @return active Number of active DIDs
     * @return revoked Number of revoked DIDs
     */
    function getStats() external view returns (uint256 total, uint256 active, uint256 revoked) {
        return (totalDIDs, activeDIDs, revokedDIDs);
    }

    /**
     * @dev Emergency function to pause all operations (only owner)
     */
    function emergencyPause() external onlyOwner {
        // This would require implementing Pausable from OpenZeppelin
        // For now, we'll just emit an event
        emit AdminChanged(owner(), address(0));
    }

    /**
     * @dev Batch register multiple DIDs
     * @param userHashes Array of user hashes
     * @param dids Array of DIDs
     * @param metadatas Array of metadata strings
     */
    function batchRegisterDIDs(
        bytes32[] calldata userHashes,
        string[] calldata dids,
        string[] calldata metadatas
    ) external onlyAuthorized nonReentrant {
        require(
            userHashes.length == dids.length && dids.length == metadatas.length,
            "DIDRegistry: array lengths must match"
        );

        for (uint256 i = 0; i < userHashes.length; i++) {
            if (didRecords[userHashes[i]].registrationTime == 0) {
                _registerDIDInternal(userHashes[i], dids[i], metadatas[i]);
            }
        }
    }
}
