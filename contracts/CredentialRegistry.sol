// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title CredentialRegistry
 * @notice 高校毕业生技能成果可信存证合约（联盟链）
 *
 * 设计思路（答辩要点）：
 *   本合约是整个系统的信任锚点。技能成果的数字指纹（文件哈希）
 *   一旦写入链上，任何人都无法篡改，从而实现了「可信存证」。
 *
 * 核心流程：
 *   ① 授权节点（学校/实训企业）调用 storeCredential() 将成果哈希上链
 *   ② 企业核验时调用 verifyCredential() 比对传入哈希与链上哈希
 *   ③ 全部操作通过 logOperation() 留痕，形成审计链条
 *
 * 防篡改原理：
 *   - 存证编号全局唯一，require 拒绝重复写入
 *   - 合约不提供 update/delete 函数，数据只增不改
 *   - 存储映射声明为 private，外部无法直接修改
 */
contract CredentialRegistry {

    // ========== 成果类型常量（与后端 models 包一致） ==========
    string constant TYPE_CERTIFICATE  = "certificate";  // 技能证书
    string constant TYPE_COMPETITION  = "competition";  // 竞赛获奖
    string constant TYPE_INTERNSHIP   = "internship";   // 岗位实训
    string constant TYPE_PROJECT      = "project";      // 项目实践

    // ========== 存证结构 ==========
    struct Credential {
        string  credentialNo;    // 唯一存证编号
        string  studentId;       // 学生ID（学号）
        string  achievementType; // 成果类型
        string  title;           // 成果名称
        string  fileHash;        // 文件 SHA-256 哈希
        uint256 timestamp;       // 上链时间戳
        address issuer;          // 上链节点地址（学校/实训企业）
        string  issuerName;      // 上链单位名称
        bool    exists;          // 标记该编号是否已存在
    }

    // ========== 状态变量 ==========
    address public owner;                       // 合约管理员（联盟链管理节点）
    mapping(address => bool) public authorizedNodes; // 授权上链节点
    mapping(address => bool) public registeredStudents; // 已注册学生地址
    mapping(string  => Credential) private credentials; // 编号 => 存证
    string[] private credentialNos;             // 全部存证编号（只追加）

    // 操作日志：仅存哈希，保护隐私
    string[] public operationLogs;

    // ========== 事件 ==========
    event StudentRegistered(address indexed studentAddr, string studentId);
    event CredentialStored(
        string  indexed credentialNo,
        string  studentId,
        string  achievementType,
        string  fileHash,
        address indexed issuer,
        uint256 timestamp
    );
    event CredentialVerified(
        string  indexed credentialNo,
        bool    passed,
        address verifier
    );
    event OperationLogged(string opHash, address operator, uint256 timestamp);

    // ========== 修饰器 ==========
    modifier onlyOwner() {
        require(msg.sender == owner, "Only contract owner");
        _;
    }

    modifier onlyAuthorized() {
        require(authorizedNodes[msg.sender], "Not authorized node");
        _;
    }

    // ========== 构造函数 ==========
    constructor() {
        owner = msg.sender;
        authorizedNodes[msg.sender] = true; // 部署者默认为授权节点
    }

    // ===================== 管理功能 =====================

    /**
     * @notice 添加授权节点（学校 / 实训企业）
     */
    function addAuthorizedNode(address nodeAddr) external onlyOwner {
        require(nodeAddr != address(0), "Invalid address");
        authorizedNodes[nodeAddr] = true;
    }

    /**
     * @notice 移除授权节点
     */
    function removeAuthorizedNode(address nodeAddr) external onlyOwner {
        authorizedNodes[nodeAddr] = false;
    }

    // ===================== 学生身份 =====================

    /**
     * @notice 学生链身份注册（学生本人或授权节点代注册）
     * @param studentId 学号
     */
    function registerStudent(string calldata studentId) external {
        require(bytes(studentId).length > 0, "Empty studentId");
        registeredStudents[msg.sender] = true;
        emit StudentRegistered(msg.sender, studentId);
    }

    // ===================== 核心：成果上链存证 =====================

    /**
     * @notice 成果存证上链（仅授权节点可调用）
     * @param credentialNo 唯一存证编号（由后端生成，如 CC-20240601-A1B2C3）
     * @param studentId    学号
     * @param achType      成果类型：certificate/competition/internship/project
     * @param title        成果名称
     * @param fileHash     附件文件 SHA-256（64位十六进制）
     * @param issuerName   上链单位名称
     * @return 成功返回 true
     *
     * 关键约束：
     *   - credentialNo 全局唯一，不可覆盖（不可篡改）
     *   - 仅授权节点（学校/实训企业）可上链
     *   - 文件哈希与元数据一旦写入不可修改
     */
    function storeCredential(
        string calldata credentialNo,
        string calldata studentId,
        string calldata achType,
        string calldata title,
        string calldata fileHash,
        string calldata issuerName
    ) external onlyAuthorized returns (bool) {
        // 唯一性校验：已存在的编号拒绝覆盖
        require(bytes(credentialNo).length > 0, "Empty credentialNo");
        require(!credentials[credentialNo].exists, "Credential already exists");
        require(bytes(studentId).length > 0, "Empty studentId");
        require(bytes(fileHash).length == 64, "FileHash must be SHA-256 hex (64 chars)");
        require(_isValidType(achType), "Invalid achievement type");

        Credential memory c = Credential({
            credentialNo:    credentialNo,
            studentId:       studentId,
            achievementType: achType,
            title:           title,
            fileHash:        fileHash,
            timestamp:       block.timestamp,
            issuer:          msg.sender,
            issuerName:      issuerName,
            exists:          true
        });

        credentials[credentialNo] = c;
        credentialNos.push(credentialNo);

        emit CredentialStored(
            credentialNo,
            studentId,
            achType,
            fileHash,
            msg.sender,
            block.timestamp
        );
        return true;
    }

    // ===================== 核心：凭证核验 =====================

    /**
     * @notice 凭证真实性核验：比对传入文件哈希与链上存证哈希
     * @param credentialNo 存证凭证编号
     * @param fileHash     待核验文件的 SHA-256
     * @return passed   核验是否通过
     * @return studentId 学号
     * @return achType  成果类型
     * @return title    成果名称
     * @return timestamp 上链时间
     * @return issuerName 上链单位
     */
    function verifyCredential(
        string calldata credentialNo,
        string calldata fileHash
    )
        external
        returns (
            bool    passed,
            string  memory studentId,
            string  memory achType,
            string  memory title,
            uint256 timestamp,
            string  memory issuerName
        )
    {
        Credential storage c = credentials[credentialNo];
        require(c.exists, "Credential not found");

        // 核心比对：文件哈希是否一致
        passed = _equals(c.fileHash, fileHash);

        emit CredentialVerified(credentialNo, passed, msg.sender);

        return (passed, c.studentId, c.achievementType, c.title, c.timestamp, c.issuerName);
    }

    /**
     * @notice 只读查询存证信息（不触发事件，供前端展示）
     */
    function getCredential(string calldata credentialNo)
        external
        view
        returns (
            string  memory studentId,
            string  memory achType,
            string  memory title,
            string  memory fileHash,
            uint256 timestamp,
            address issuer,
            string  memory issuerName,
            bool    exists
        )
    {
        Credential storage c = credentials[credentialNo];
        return (
            c.studentId,
            c.achievementType,
            c.title,
            c.fileHash,
            c.timestamp,
            c.issuer,
            c.issuerName,
            c.exists
        );
    }

    /**
     * @notice 按学号查询该学生全部存证编号
     * @dev 遍历实现，演示规模足够；生产可换索引结构
     */
    function getCredentialNosByStudent(string calldata studentId)
        external
        view
        returns (string[] memory nos)
    {
        uint256 count = 0;
        for (uint256 i = 0; i < credentialNos.length; i++) {
            if (_equals(credentials[credentialNos[i]].studentId, studentId)) {
                count++;
            }
        }
        nos = new string[](count);
        uint256 idx = 0;
        for (uint256 i = 0; i < credentialNos.length; i++) {
            if (_equals(credentials[credentialNos[i]].studentId, studentId)) {
                nos[idx++] = credentialNos[i];
            }
        }
    }

    /**
     * @notice 存证总数
     */
    function totalCredentials() external view returns (uint256) {
        return credentialNos.length;
    }

    // ===================== 操作日志上链 =====================

    /**
     * @notice 操作日志哈希上链（全部存证提交、外部核验访问均需记录）
     * @param opHash 操作日志的 SHA-256 哈希（不存原文，保护隐私）
     */
    function logOperation(string calldata opHash) external onlyAuthorized {
        require(bytes(opHash).length == 64, "opHash must be SHA-256 hex");
        operationLogs.push(opHash);
        emit OperationLogged(opHash, msg.sender, block.timestamp);
    }

    /**
     * @notice 查询日志条数
     */
    function totalOperationLogs() external view returns (uint256) {
        return operationLogs.length;
    }

    // ===================== 内部工具 =====================

    function _isValidType(string calldata achType) internal pure returns (bool) {
        return _equals(achType, TYPE_CERTIFICATE)
            || _equals(achType, TYPE_COMPETITION)
            || _equals(achType, TYPE_INTERNSHIP)
            || _equals(achType, TYPE_PROJECT);
    }

    /// @dev 字符串比较（避免 keccak 跨版本差异，演示环境足够）
    function _equals(string memory a, string memory b) internal pure returns (bool) {
        return keccak256(abi.encodePacked(a)) == keccak256(abi.encodePacked(b));
    }
}
