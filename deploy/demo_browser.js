// 演示脚本 v3：修复选择器，慢速全屏
// 环境变量：DEMO_BASE_URL（默认 http://127.0.0.1）、DEMO_PASSWORD（演示账号口令，必填）
const { chromium } = require('playwright');

const BASE_URL = process.env.DEMO_BASE_URL || 'http://127.0.0.1';
const DEMO_CRED_NO = process.env.DEMO_CRED_NO || 'CC-20260927120316-A3D0B3';
const DEMO_HASH = process.env.DEMO_HASH || 'aa193a9593bf51cff846bc74f8a01d9a5ccc8103ce042ec4f44df72fd8520bb8';
const DEMO_PASSWORD = process.env.DEMO_PASSWORD;
if (!DEMO_PASSWORD) {
    console.error('缺少演示口令：请通过环境变量 DEMO_PASSWORD 提供（见 deploy/.env.demo）');
    process.exit(1);
}

(async () => {
  const browser = await chromium.launch({
    channel: 'msedge',
    headless: false,
    slowMo: 1000,
  });

  try {
    const context = await browser.newContext({
      viewport: { width: 1920, height: 1080 },
      locale: 'zh-CN',
    });
    const page = await context.newPage();
    const pause = (ms) => page.waitForTimeout(ms);

    // 【1】打开首页
    console.log('【1】打开系统首页');
    await page.goto(BASE_URL);
    await pause(4000);

    // 【2】注册
    console.log('【2】注册新账号');
    await page.click('a:has-text("注册账号")');
    await pause(3000);
    await page.locator('input').first().fill('student02');
    await pause(1500);
    await page.locator('input[type="password"]').first().fill(DEMO_PASSWORD);
    await pause(1500);
    await page.locator('.el-form-item').nth(2).locator('input').fill('张同学');
    await pause(1500);
    await page.locator('.el-select').first().click();
    await pause(2000);
    await page.locator('.el-select-dropdown__item:has-text("学生")').first().click();
    await pause(2000);
    await page.locator('input[placeholder*="学生必填"]').fill('2024002');
    await pause(1500);
    await page.locator('button:has-text("注 册")').click();
    await pause(5000);
    await page.locator('button:has-text("去登录")').click();
    await pause(3000);

    // 【3】登录管理员
    console.log('【3】登录管理员');
    await page.locator('input').first().fill('admin01');
    await pause(1500);
    await page.locator('input[type="password"]').first().fill(DEMO_PASSWORD);
    await pause(1500);
    await page.locator('button:has-text("登 录")').click();
    await pause(4000);

    // 【4】成果录入
    console.log('【4】进入成果录入');
    await page.locator('span:has-text("成果录入")').click();
    await pause(3500);
    await page.locator('button:has-text("录入成果")').click();
    await pause(3000);
    await page.locator('input[placeholder*="学号"]').fill('2024002');
    await pause(1500);
    await page.locator('button:has-text("查询")').click();
    await pause(2500);
    await page.locator('input').nth(2).fill('全国职业院校技能大赛一等奖');
    await pause(1500);
    await page.locator('input').nth(3).fill('教育部');
    await pause(1500);
    await page.locator('button:has-text("选择文件")').click();
    await pause(8000); // 手动选文件，展示哈希计算
    await page.locator('button:has-text("提交")').click();
    await pause(3000);

    // 【5】上链
    console.log('【5】上链存证');
    await page.locator('button:has-text("上链存证")').first().click();
    await pause(3000);
    await page.locator('button:has-text("确定")').click();
    await pause(8000); // 展示存证编号/交易哈希/区块高度
    await page.locator('button:has-text("确定")').click();
    await pause(2000);

    // 【6】学生查看
    console.log('【6】学生查看存证');
    await page.locator('button:has-text("退出")').click();
    await pause(2500);
    await page.locator('input').first().fill('demo_student');
    await pause(1500);
    await page.locator('input[type="password"]').first().fill(DEMO_PASSWORD);
    await pause(1500);
    await page.locator('button:has-text("登 录")').click();
    await pause(4000);
    await page.locator('span:has-text("我的存证")').click();
    await pause(4000);
    await page.locator('button:has-text("二维码")').first().click();
    await pause(5000); // 展示二维码

    // 【7】核验
    console.log('【7】打开核验页');
    await page.goto(`${BASE_URL}/verify`);
    await pause(3500);

    // 【8】核验通过
    console.log('【8】核验通过');
    await page.locator('input').first().fill(DEMO_CRED_NO);
    await pause(2500);
    // file_hash 是 textarea，不是 input
    await page.locator('textarea').first().fill(DEMO_HASH);
    await pause(2500);
    await page.locator('button:has-text("开始核验")').click();
    await pause(6000);

    // 【9】篡改模拟
    console.log('【9】篡改核验');
    await page.locator('button:has-text("继续核验")').click();
    await pause(2500);
    await page.locator('textarea').first().fill(DEMO_HASH.slice(0, -1) + '9');
    await pause(2500);
    await page.locator('button:has-text("开始核验")').click();
    await pause(6000);

    console.log('=== 演示完成！ ===');
    await pause(8000);
    await browser.close();
  } catch (e) {
    console.error('Error:', e.message);
    process.exitCode = 1;
    // 出错时保持浏览器画面便于排查，由用户手动关闭
  }
})();
