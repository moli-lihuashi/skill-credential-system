# -*- coding: utf-8 -*-
"""
答辩PPT生成：基于联盟链的高校毕业生技能成果可信存证系统
"""
import os
from pptx import Presentation
from pptx.util import Inches, Pt, Emu
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN
from pptx.enum.shapes import MSO_SHAPE
from pptx.oxml.ns import qn

OUTPUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '答辩PPT-技能成果可信存证系统.pptx')

prs = Presentation()
prs.slide_width = Inches(13.333)
prs.slide_height = Inches(7.5)

# 配色
PRIMARY = RGBColor(0x1A, 0x3A, 0x5C)
ACCENT = RGBColor(0x18, 0x90, 0xFF)
WHITE = RGBColor(0xFF, 0xFF, 0xFF)
DARK = RGBColor(0x30, 0x31, 0x33)
GRAY = RGBColor(0x90, 0x93, 0x99)
LIGHT_BG = RGBColor(0xF5, 0xF7, 0xFA)
SUCCESS = RGBColor(0x52, 0xC4, 0x1A)
WARNING = RGBColor(0xFA, 0x8C, 0x16)

def add_bg(slide, color):
    bg = slide.background
    fill = bg.fill
    fill.solid()
    fill.fore_color.rgb = color

def add_rect(slide, left, top, width, height, color):
    shape = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE, left, top, width, height)
    shape.fill.solid()
    shape.fill.fore_color.rgb = color
    shape.line.fill.background()
    return shape

def add_round_rect(slide, left, top, width, height, color):
    shape = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, left, top, width, height)
    shape.fill.solid()
    shape.fill.fore_color.rgb = color
    shape.line.fill.background()
    return shape

def add_text(slide, left, top, width, height, text, size=18, bold=False, color=DARK, align=PP_ALIGN.LEFT, font='微软雅黑'):
    txBox = slide.shapes.add_textbox(left, top, width, height)
    tf = txBox.text_frame
    tf.word_wrap = True
    p = tf.paragraphs[0]
    p.text = text
    p.font.size = Pt(size)
    p.font.bold = bold
    p.font.color.rgb = color
    p.font.name = font
    p.alignment = align
    # CJK font
    for run in p.runs:
        rPr = run._r.get_or_add_rPr()
        ea = rPr.makeelement(qn('a:ea'), {})
        ea.set('typeface', font)
        rPr.append(ea)
    return txBox

def add_bullet_text(slide, left, top, width, height, items, size=16, color=DARK, spacing=1.3):
    txBox = slide.shapes.add_textbox(left, top, width, height)
    tf = txBox.text_frame
    tf.word_wrap = True
    for i, item in enumerate(items):
        if i == 0:
            p = tf.paragraphs[0]
        else:
            p = tf.add_paragraph()
        p.text = item
        p.font.size = Pt(size)
        p.font.color.rgb = color
        p.font.name = '微软雅黑'
        p.space_after = Pt(6)
        p.line_spacing = spacing
        for run in p.runs:
            rPr = run._r.get_or_add_rPr()
            ea = rPr.makeelement(qn('a:ea'), {})
            ea.set('typeface', '微软雅黑')
            rPr.append(ea)
    return txBox

def slide_header(slide, title, subtitle=''):
    add_rect(slide, Inches(0), Inches(0), Inches(13.333), Inches(1.1), PRIMARY)
    add_text(slide, Inches(0.5), Inches(0.15), Inches(10), Inches(0.6), title, size=28, bold=True, color=WHITE)
    if subtitle:
        add_text(slide, Inches(0.5), Inches(0.65), Inches(10), Inches(0.4), subtitle, size=14, color=RGBColor(0xBB, 0xCC, 0xDD))

# ============================================================
# Slide 1: 封面
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_rect(slide, Inches(0), Inches(0), Inches(13.333), Inches(7.5), PRIMARY)
add_rect(slide, Inches(0), Inches(5.8), Inches(13.333), Inches(0.06), ACCENT)

add_text(slide, Inches(1), Inches(1.2), Inches(11), Inches(0.8),
         '＿＿＿＿＿＿＿＿＿＿', size=24, color=RGBColor(0x88, 0xAA, 0xCC), align=PP_ALIGN.CENTER)
add_text(slide, Inches(1), Inches(1.8), Inches(11), Inches(0.8),
         '＿＿＿＿＿＿学院 · ＿＿＿＿＿＿专业', size=16, color=RGBColor(0x88, 0xAA, 0xCC), align=PP_ALIGN.CENTER)

add_text(slide, Inches(1), Inches(2.8), Inches(11), Inches(1.2),
         '基于联盟链的高校毕业生\n技能成果可信存证系统', size=38, bold=True, color=WHITE, align=PP_ALIGN.CENTER)

add_text(slide, Inches(1), Inches(4.5), Inches(11), Inches(0.5),
         '毕业设计（论文）答辩', size=20, color=RGBColor(0xBB, 0xCC, 0xDD), align=PP_ALIGN.CENTER)

add_text(slide, Inches(1), Inches(6.1), Inches(11), Inches(0.8),
         '指导教师：＿＿＿＿    小组成员：＿＿＿＿（4人）\n答辩人：＿＿＿＿', size=16, color=RGBColor(0x88, 0xAA, 0xCC), align=PP_ALIGN.CENTER)

# ============================================================
# Slide 2: 目录
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '答辩内容提纲', 'CONTENTS')

items = [
    ('01', '项目需求与背景', '痛点分析、建设目标'),
    ('02', '小组分工', '四人模块划分与协作'),
    ('03', '系统架构设计', '四层架构、链上链下划分'),
    ('04', '核心功能与技术实现', '智能合约、后端、前端'),
    ('05', '系统测试与演示', '功能测试、篡改模拟'),
    ('06', '总结与展望', '创新点、不足与改进'),
]
for i, (num, title, desc) in enumerate(items):
    row = i // 3
    col = i % 3
    x = Inches(0.8 + col * 4.0)
    y = Inches(1.8 + row * 2.5)
    add_round_rect(slide, x, y, Inches(3.6), Inches(2.0), WHITE)
    add_text(slide, x + Inches(0.3), y + Inches(0.2), Inches(3), Inches(0.5), num, size=28, bold=True, color=ACCENT)
    add_text(slide, x + Inches(0.3), y + Inches(0.8), Inches(3), Inches(0.5), title, size=18, bold=True, color=DARK)
    add_text(slide, x + Inches(0.3), y + Inches(1.3), Inches(3), Inches(0.5), desc, size=13, color=GRAY)

# ============================================================
# Slide 3: 项目需求
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '01  项目需求与背景', '为什么要做这个系统？')

# 左侧：痛点
add_round_rect(slide, Inches(0.5), Inches(1.5), Inches(5.8), Inches(5.5), WHITE)
add_text(slide, Inches(0.8), Inches(1.6), Inches(5), Inches(0.5), '现实痛点', size=20, bold=True, color=DARK)
add_bullet_text(slide, Inches(0.8), Inches(2.2), Inches(5.2), Inches(4.5), [
    '•  纸质证书容易伪造，核验成本高',
    '•  用人单位难以快速辨别成果真伪',
    '•  纸质材料核验流程繁琐、效率低',
    '•  成果存证缺乏权威第三方背书',
    '•  高校毕业生求职时成果认证难',
], size=15, color=DARK)

# 右侧：目标
add_round_rect(slide, Inches(6.8), Inches(1.5), Inches(6), Inches(5.5), WHITE)
add_text(slide, Inches(7.1), Inches(1.6), Inches(5), Inches(0.5), '建设目标', size=20, bold=True, color=ACCENT)
add_bullet_text(slide, Inches(7.1), Inches(2.2), Inches(5.5), Inches(4.5), [
    '✦  搭建联盟链可信存证平台',
    '✦  学校/实训企业作为授权节点',
    '✦  技能成果哈希上链，不可篡改',
    '✦  企业输入编号即可在线核验',
    '✦  链上只存哈希，保护学生隐私',
], size=15, color=DARK)

# ============================================================
# Slide 4: 小组分工
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '02  小组分工', '4人协作 · 各自承担实质性任务')

roles = [
    ('同学A\n架构/部署', '系统架构设计\n需求分析\nLinux部署联调', ACCENT),
    ('同学B\n智能合约', 'Solidity合约开发\n合约安全设计\n链上事件设计', RGBColor(0x2E, 0xCC, 0x71)),
    ('同学C\n后端开发', 'Golang服务\nAPI设计实现\n链交互/哈希/日志', WARNING),
    ('同学D\n前端开发', 'Vue3四端界面\n路由权限控制\n核验交互/演示', RGBColor(0xE7, 0x4C, 0x3C)),
]
for i, (role, detail, color) in enumerate(roles):
    x = Inches(0.5 + i * 3.2)
    add_round_rect(slide, x, Inches(1.6), Inches(2.9), Inches(5.2), WHITE)
    # 头部色块
    add_rect(slide, x, Inches(1.6), Inches(2.9), Inches(1.6), color)
    add_text(slide, x + Inches(0.2), Inches(1.8), Inches(2.5), Inches(1.3), role, size=18, bold=True, color=WHITE, align=PP_ALIGN.CENTER)
    add_bullet_text(slide, x + Inches(0.3), Inches(3.5), Inches(2.4), Inches(3),
                    [f'•  {line}' for line in detail.split('\n')], size=13, color=DARK)

# ============================================================
# Slide 5: 系统架构
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '03  系统架构设计', '四层架构 · 链上存哈希、链下存原文')

layers = [
    ('展示层 (Vue3)', '学生端 │ 院校管理端 │ 实训企业端 │ 核验端', ACCENT),
    ('业务层 (Golang)', '账号权限 │ 成果管理 │ 存证服务 │ 核验服务 │ 日志服务', RGBColor(0x2E, 0xCC, 0x71)),
    ('联盟链层 (Solidity)', 'CredentialRegistry 合约：存证 / 核验 / 日志上链', WARNING),
    ('数据层 (MySQL + IPFS)', '用户表 │ 成果表 │ 凭证表 │ 核验日志 │ 系统日志', RGBColor(0xE7, 0x4C, 0x3C)),
]
for i, (title, desc, color) in enumerate(layers):
    y = Inches(1.5 + i * 1.45)
    add_round_rect(slide, Inches(1.5), y, Inches(10.3), Inches(1.2), WHITE)
    add_rect(slide, Inches(1.5), y, Inches(0.15), Inches(1.2), color)
    add_text(slide, Inches(2.0), y + Inches(0.1), Inches(3), Inches(0.5), title, size=18, bold=True, color=color)
    add_text(slide, Inches(5.2), y + Inches(0.15), Inches(6.3), Inches(0.8), desc, size=14, color=DARK)

# 箭头说明
add_text(slide, Inches(1.5), Inches(7.0), Inches(10), Inches(0.4),
         'HTTP/JWT ←→  ethclient/ABI ←→  SQL', size=12, color=GRAY, align=PP_ALIGN.CENTER)

# ============================================================
# Slide 6: 数据库设计
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '03  数据库设计', 'MySQL 5张核心表')

tables = [
    ('users 用户表', 'username, password(bcrypt), role, eth_address, eth_private_key', '注册自动生成链账户'),
    ('achievements 成果表', 'title, ach_type, file_hash, status, created_by', '4类成果：证书/竞赛/实训/项目'),
    ('credentials 凭证表', 'credential_no(唯一), chain_tx_hash, file_hash', '只增不改删！对应链上不可篡改'),
    ('verify_logs 核验表', 'credential_no, employer_id, input_hash, chain_result', '每次核验留痕'),
    ('sys_logs 日志表', 'action, detail, log_hash, onchain', '日志哈希上链留痕'),
]
for i, (name, fields, note) in enumerate(tables):
    y = Inches(1.4 + i * 1.15)
    add_round_rect(slide, Inches(0.5), y, Inches(12.3), Inches(1.0), WHITE)
    add_text(slide, Inches(0.8), y + Inches(0.08), Inches(3.2), Inches(0.4), name, size=14, bold=True, color=PRIMARY)
    add_text(slide, Inches(0.8), y + Inches(0.5), Inches(8.5), Inches(0.4), fields, size=11, color=GRAY)
    add_text(slide, Inches(9.5), y + Inches(0.25), Inches(3.2), Inches(0.5), note, size=12, bold=True, color=ACCENT)

# ============================================================
# Slide 7: 智能合约
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '04  智能合约设计与实现', 'CredentialRegistry.sol')

# 左侧：核心接口
add_round_rect(slide, Inches(0.5), Inches(1.5), Inches(6), Inches(5.5), WHITE)
add_text(slide, Inches(0.8), Inches(1.6), Inches(5), Inches(0.5), '核心合约接口', size=18, bold=True, color=DARK)
add_bullet_text(slide, Inches(0.8), Inches(2.2), Inches(5.5), Inches(4.5), [
    '▸  storeCredential() — 成果上链存证',
    '     参数: 编号/学号/类型/哈希/单位',
    '▸  verifyCredential() — 凭证核验',
    '     比对哈希 → 返回通过/不通过',
    '▸  getCredential() — 只读查询',
    '▸  logOperation() — 日志哈希上链',
    '▸  addAuthorizedNode() — 授权管理',
], size=14, color=DARK)

# 右侧：防篡改
add_round_rect(slide, Inches(7), Inches(1.5), Inches(5.8), Inches(5.5), WHITE)
add_text(slide, Inches(7.3), Inches(1.6), Inches(5), Inches(0.5), '不可篡改机制', size=18, bold=True, color=ACCENT)
add_bullet_text(slide, Inches(7.3), Inches(2.2), Inches(5.3), Inches(4.5), [
    '✦  存证编号全局唯一，拒绝覆盖写入',
    '✦  合约无 update/delete 函数',
    '✦  credentials 映射私有存储',
    '✦  全部操作触发 Event 事件',
    '',
    '核验原理：',
    '  SHA-256(文件) == 链上存储哈希？',
    '  ✓ 一致 → 核验通过（数据未篡改）',
    '  ✗ 不一致 → 核验不通过（篡改风险）',
], size=14, color=DARK)

# ============================================================
# Slide 8: 后端技术实现
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '04  后端技术实现', 'Golang + Gin + GORM + go-ethereum')

items = [
    ('JWT 鉴权', '登录签发 HS256 令牌\n中间件拦截验证\nRoleRequired 角色控制'),
    ('链账户生成', '注册自动分配链账户\nECDSA(secp256k1) 密钥\n每个用户独立地址'),
    ('文件哈希', 'SHA-256 流式计算\n上传附件自动算哈希\n哈希值写入合约'),
    ('链上交互', 'go-ethereum ABI 编码\nEIP-155 交易签名\n等待回执确认'),
]
for i, (title, desc) in enumerate(items):
    x = Inches(0.5 + i * 3.2)
    add_round_rect(slide, x, Inches(1.5), Inches(2.9), Inches(3.5), WHITE)
    add_rect(slide, x, Inches(1.5), Inches(2.9), Inches(0.8), PRIMARY)
    add_text(slide, x + Inches(0.2), Inches(1.6), Inches(2.5), Inches(0.6), title, size=16, bold=True, color=WHITE, align=PP_ALIGN.CENTER)
    add_bullet_text(slide, x + Inches(0.3), Inches(2.5), Inches(2.4), Inches(2.3),
                    [f'•  {l}' for l in desc.split('\n')], size=12, color=DARK)

# API 简表
add_round_rect(slide, Inches(0.5), Inches(5.3), Inches(12.3), Inches(1.8), WHITE)
add_text(slide, Inches(0.8), Inches(5.4), Inches(5), Inches(0.4), '核心 API 接口', size=14, bold=True, color=PRIMARY)
add_text(slide, Inches(0.8), Inches(5.85), Inches(12), Inches(1.0),
         'POST /auth/register    POST /auth/login    POST /achievement    POST /credential/store\n'
         'GET  /credential/my    GET /credential/search    POST /verify    GET /logs',
         size=12, color=DARK, font='Consolas')

# ============================================================
# Slide 9: 前端界面
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '04  前端界面实现', 'Vue3 + Element Plus · 四端独立界面')

pages = [
    ('学生端', '个人中心 · 我的存证', '查看链上存证凭证\n存证概览统计', ACCENT),
    ('院校管理端', '工作台 · 成果录入 · 存证管理 · 日志', '录入证书/竞赛\n上链存证 · 检索 · 日志', RGBColor(0x2E, 0xCC, 0x71)),
    ('实训企业端', '工作台 · 实训成果录入', '录入实训评价/项目\n上链存证', WARNING),
    ('招聘企业核验端', '凭证核验 · 核验历史', '输入编号核验\n查看核验记录', RGBColor(0xE7, 0x4C, 0x3C)),
]
for i, (title, nav, desc, color) in enumerate(pages):
    x = Inches(0.5 + i * 3.2)
    add_round_rect(slide, x, Inches(1.5), Inches(2.9), Inches(5.3), WHITE)
    add_rect(slide, x, Inches(1.5), Inches(2.9), Inches(1.0), color)
    add_text(slide, x + Inches(0.2), Inches(1.65), Inches(2.5), Inches(0.7), title, size=16, bold=True, color=WHITE, align=PP_ALIGN.CENTER)
    add_text(slide, x + Inches(0.3), Inches(2.7), Inches(2.4), Inches(0.8), nav, size=11, color=GRAY)
    add_bullet_text(slide, x + Inches(0.3), Inches(3.5), Inches(2.4), Inches(2.5),
                    [f'•  {l}' for l in desc.split('\n')], size=13, color=DARK)

# ============================================================
# Slide 10: 核心业务流程
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '04  核心业务流程', '成果上传 → 存证 → 企业核验')

steps = [
    ('1', '录入成果', '院校/企业录入\n上传附件'),
    ('2', '计算哈希', 'SHA-256\n文件数字指纹'),
    ('3', '上链存证', '调用合约\n返回存证编号'),
    ('4', '生成凭证', '唯一凭证编号\nCC-时间-随机'),
    ('5', '企业核验', '输入编号+哈希\n比对链上数据'),
    ('6', '输出结果', '通过/不通过\n篡改风险提示'),
]
for i, (num, title, desc) in enumerate(steps):
    x = Inches(0.3 + i * 2.15)
    add_round_rect(slide, x, Inches(1.8), Inches(1.95), Inches(3.2), WHITE)
    circle = slide.shapes.add_shape(MSO_SHAPE.OVAL, x + Inches(0.65), Inches(2.0), Inches(0.65), Inches(0.65))
    circle.fill.solid()
    circle.fill.fore_color.rgb = ACCENT
    circle.line.fill.background()
    add_text(slide, x + Inches(0.65), Inches(2.05), Inches(0.65), Inches(0.55), num, size=18, bold=True, color=WHITE, align=PP_ALIGN.CENTER)
    add_text(slide, x + Inches(0.1), Inches(2.8), Inches(1.8), Inches(0.5), title, size=14, bold=True, color=DARK, align=PP_ALIGN.CENTER)
    add_text(slide, x + Inches(0.1), Inches(3.3), Inches(1.8), Inches(1.2), desc, size=11, color=GRAY, align=PP_ALIGN.CENTER)
    # 箭头
    if i < 5:
        add_text(slide, x + Inches(1.9), Inches(2.8), Inches(0.3), Inches(0.5), '→', size=20, color=ACCENT)

# 底部说明
add_round_rect(slide, Inches(0.5), Inches(5.5), Inches(12.3), Inches(1.5), WHITE)
add_text(slide, Inches(0.8), Inches(5.6), Inches(11.5), Inches(1.2),
         '核验场景演示：\n'
         '① 使用正确文件哈希核验 → 输出"核验通过（数据未篡改）"\n'
         '② 修改文件内容后重新核验 → 输出"核验不通过（信息存在篡改风险）" ← 篡改模拟测试',
         size=14, color=DARK)

# ============================================================
# Slide 11: 测试结果
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '05  系统测试', '功能测试 + 篡改模拟测试 + 并发测试')

tests = [
    ('用户注册测试', 'PASS', '四类角色注册成功，自动生成链账户'),
    ('登录鉴权测试', 'PASS', 'JWT签发验证正常，角色权限控制有效'),
    ('成果录入测试', 'PASS', '文件哈希计算准确，角色权限隔离'),
    ('上链存证测试', 'PASS', '合约调用成功，存证编号唯一'),
    ('凭证核验测试', 'PASS', '通过/不通过输出正确'),
    ('篡改模拟测试', 'PASS', '修改文件后核验不通过 ✓ 安全机制有效'),
    ('并发核验测试', 'PASS', '10并发请求全部正常响应'),
]
for i, (name, result, desc) in enumerate(tests):
    y = Inches(1.4 + i * 0.82)
    add_round_rect(slide, Inches(0.5), y, Inches(12.3), Inches(0.72), WHITE)
    tag = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, Inches(0.7), y + Inches(0.12), Inches(0.9), Inches(0.48))
    tag.fill.solid()
    tag.fill.fore_color.rgb = SUCCESS
    tag.line.fill.background()
    add_text(slide, Inches(0.7), y + Inches(0.12), Inches(0.9), Inches(0.48), result, size=11, bold=True, color=WHITE, align=PP_ALIGN.CENTER)
    add_text(slide, Inches(1.8), y + Inches(0.12), Inches(3), Inches(0.48), name, size=14, bold=True, color=DARK)
    add_text(slide, Inches(5), y + Inches(0.12), Inches(7.5), Inches(0.48), desc, size=13, color=GRAY)

# ============================================================
# Slide 12: 总结
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_bg(slide, LIGHT_BG)
slide_header(slide, '06  总结与展望', '创新点 · 不足 · 未来方向')

# 创新点
add_round_rect(slide, Inches(0.5), Inches(1.5), Inches(6), Inches(2.8), WHITE)
add_text(slide, Inches(0.8), Inches(1.6), Inches(5), Inches(0.5), '✦ 主要创新点', size=18, bold=True, color=ACCENT)
add_bullet_text(slide, Inches(0.8), Inches(2.1), Inches(5.5), Inches(2.0), [
    '•  面向职业教育四类成果的联盟链存证',
    '•  多角色协作权限控制与业务流程',
    '•  SHA-256哈希比对防篡改核验机制',
    '•  操作日志哈希上链留痕审计',
], size=14, color=DARK)

# 不足与展望
add_round_rect(slide, Inches(7), Inches(1.5), Inches(5.8), Inches(2.8), WHITE)
add_text(slide, Inches(7.3), Inches(1.6), Inches(5), Inches(0.5), '▸ 不足与展望', size=18, bold=True, color=WARNING)
add_bullet_text(slide, Inches(7.3), Inches(2.1), Inches(5.3), Inches(2.0), [
    '•  后续部署多节点联盟链网络',
    '•  集成IPFS去中心化文件存储',
    '•  引入零知识证明保护隐私',
    '•  开发移动端（小程序）版本',
], size=14, color=DARK)

# 底部技术栈
add_round_rect(slide, Inches(0.5), Inches(4.7), Inches(12.3), Inches(2.3), WHITE)
add_text(slide, Inches(0.8), Inches(4.85), Inches(5), Inches(0.5), '技术栈总览', size=18, bold=True, color=PRIMARY)
add_text(slide, Inches(0.8), Inches(5.4), Inches(11.5), Inches(1.3),
         '智能合约：Solidity 0.8.19     后端：Golang 1.21 + Gin + GORM + go-ethereum\n'
         '前端：Vue3 + Element Plus + Vue Router + Pinia     数据库：MySQL 5.7\n'
         '区块链：Ganache 本地联盟链（可扩展私链）     部署：Linux CentOS7 + Nginx',
         size=14, color=DARK)

# ============================================================
# Slide 13: 致谢
# ============================================================
slide = prs.slides.add_slide(prs.slide_layouts[6])
add_rect(slide, Inches(0), Inches(0), Inches(13.333), Inches(7.5), PRIMARY)
add_rect(slide, Inches(0), Inches(5.8), Inches(13.333), Inches(0.06), ACCENT)

add_text(slide, Inches(1), Inches(2.0), Inches(11), Inches(1.2),
         '感谢聆听', size=48, bold=True, color=WHITE, align=PP_ALIGN.CENTER)
add_text(slide, Inches(1), Inches(3.5), Inches(11), Inches(0.8),
         '敬请各位老师批评指正', size=24, color=RGBColor(0xBB, 0xCC, 0xDD), align=PP_ALIGN.CENTER)
add_text(slide, Inches(1), Inches(5.0), Inches(11), Inches(0.5),
         '基于联盟链的高校毕业生技能成果可信存证系统', size=14, color=RGBColor(0x88, 0xAA, 0xCC), align=PP_ALIGN.CENTER)

# 保存
prs.save(OUTPUT)
print(f'答辩PPT完成: {OUTPUT}')
print(f'共 {len(prs.slides)} 页')
