/*
@Author: omenkk7
@Date: 2024/10/26
@Description: 生成测试数据脚本 - 为所有集合生成完整的测试数据
*/

// 连接到数据库
const db = db.getSiblingDB('zk_code_arena');

print('🚀 开始生成测试数据...');

// ==================== 1. 用户数据 ====================
print('📝 生成用户数据...');

const users = [
    {
        _id: ObjectId(),
        username: "admin",
        password: "$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8imdVMaM7ZX/W3xGD.7xUlT8r2.Uy", // 密码: admin123
        email: "admin@zkcodearena.com",
        real_name: "系统管理员",
        student_id: "ADMIN001",
        role: "admin",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=admin",
        bio: "系统管理员，负责平台维护和管理",
        school: "xx大学",
        major: "计算机科学与技术",
        grade: "2024",
        class: "管理员",
        phone: "13800138000",
        is_active: true,
        last_login_at: new Date(),
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        username: "teacher1",
        password: "$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8imdVMaM7ZX/W3xGD.7xUlT8r2.Uy", // 密码: admin123
        email: "teacher1@zkcodearena.com",
        real_name: "张教授",
        student_id: "T001",
        role: "teacher",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=teacher1",
        bio: "算法与数据结构课程教师",
        school: "xx大学",
        major: "计算机科学与技术",
        grade: "教师",
        class: "计算机学院",
        phone: "13800138001",
        is_active: true,
        last_login_at: new Date(Date.now() - 3600000), // 1小时前
        created_at: new Date(Date.now() - 86400000 * 30), // 30天前
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        username: "teacher2",
        password: "$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8imdVMaM7ZX/W3xGD.7xUlT8r2.Uy",
        email: "teacher2@zkcodearena.com",
        real_name: "李老师",
        student_id: "T002",
        role: "teacher",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=teacher2",
        bio: "程序设计基础课程教师",
        school: "xx大学",
        major: "软件工程",
        grade: "教师",
        class: "软件学院",
        phone: "13800138002",
        is_active: true,
        last_login_at: new Date(Date.now() - 7200000), // 2小时前
        created_at: new Date(Date.now() - 86400000 * 25),
        updated_at: new Date()
    }
];

// 生成学生用户
const studentNames = ["王小明", "李小红", "张三", "李四", "王五", "赵六", "孙七", "周八", "吴九", "郑十"];
const majors = ["计算机科学与技术", "软件工程", "网络工程", "信息安全", "数据科学与大数据技术"];
const classes = ["计科1班", "计科2班", "软工1班", "软工2班", "网工1班"];

for (let i = 0; i < 20; i++) {
    const studentId = `2024${String(i + 1).padStart(4, '0')}`;
    users.push({
        _id: ObjectId(),
        username: `student${i + 1}`,
        password: "$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8imdVMaM7ZX/W3xGD.7xUlT8r2.Uy", // 密码: admin123
        email: `student${i + 1}@zkcodearena.com`,
        real_name: studentNames[i % studentNames.length] + (i > 9 ? (i - 9) : ""),
        student_id: studentId,
        role: "student",
        avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=student${i + 1}`,
        bio: `我是${studentNames[i % studentNames.length]}，热爱编程！`,
        school: "xx大学",
        major: majors[i % majors.length],
        grade: "2024",
        class: classes[i % classes.length],
        phone: `138${String(i + 1).padStart(8, '0')}`,
        is_active: true,
        last_login_at: new Date(Date.now() - Math.random() * 86400000 * 7), // 随机7天内
        created_at: new Date(Date.now() - 86400000 * (30 - i)), // 递减创建时间
        updated_at: new Date()
    });
}

db.users.insertMany(users);
print(`✅ 插入 ${users.length} 个用户`);

// 获取用户ID用于后续引用
const adminUser = db.users.findOne({username: "admin"});
const teacher1 = db.users.findOne({username: "teacher1"});
const teacher2 = db.users.findOne({username: "teacher2"});
const students = db.users.find({role: "student"}).toArray();

// ==================== 2. 题目数据 ====================
print('📚 生成题目数据...');

const problems = [
    {
        _id: ObjectId(),
        title: "两数之和",
        description: "给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值 target 的那两个整数，并返回它们的数组下标。\n\n你可以假设每种输入只会对应一个答案。但是，数组中同一个元素在答案里不能重复出现。\n\n你可以按任意顺序返回答案。",
        input: "第一行包含一个整数 n，表示数组长度。\n第二行包含 n 个整数，表示数组 nums。\n第三行包含一个整数 target。",
        output: "输出两个整数，表示和为 target 的两个数的下标（从0开始）。",
        sample_input: "4\n2 7 11 15\n9",
        sample_output: "0 1",
        hint: "可以使用哈希表来优化时间复杂度。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "easy",
        time_limit: 1000,
        memory_limit: 256,
        tags: ["数组", "哈希表"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher1._id,
        created_at: new Date(Date.now() - 86400000 * 20),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "回文数",
        description: "给你一个整数 x ，如果 x 是一个回文整数，返回 true ；否则，返回 false 。\n\n回文数是指正序（从左向右）和倒序（从右向左）读都是一样的整数。",
        input: "一个整数 x",
        output: "如果是回文数输出 true，否则输出 false",
        sample_input: "121",
        sample_output: "true",
        hint: "考虑负数的情况，以及如何在不转换为字符串的情况下解决。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "easy",
        time_limit: 1000,
        memory_limit: 256,
        tags: ["数学"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher1._id,
        created_at: new Date(Date.now() - 86400000 * 19),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "最长公共前缀",
        description: "编写一个函数来查找字符串数组中的最长公共前缀。\n\n如果不存在公共前缀，返回空字符串 \"\"。",
        input: "第一行包含一个整数 n，表示字符串数组的长度。\n接下来 n 行，每行包含一个字符串。",
        output: "输出最长公共前缀，如果不存在则输出空字符串。",
        sample_input: "3\nflower\nflow\nflight",
        sample_output: "fl",
        hint: "可以使用纵向扫描或横向扫描的方法。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "easy",
        time_limit: 1000,
        memory_limit: 256,
        tags: ["字符串"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher2._id,
        created_at: new Date(Date.now() - 86400000 * 18),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "三数之和",
        description: "给你一个包含 n 个整数的数组 nums，判断 nums 中是否存在三个元素 a，b，c ，使得 a + b + c = 0 ？请你找出所有和为 0 且不重复的三元组。\n\n注意：答案中不可以包含重复的三元组。",
        input: "第一行包含一个整数 n，表示数组长度。\n第二行包含 n 个整数。",
        output: "输出所有和为0的三元组，每个三元组占一行。",
        sample_input: "6\n-1 0 1 2 -1 -4",
        sample_output: "-1 -1 2\n-1 0 1",
        hint: "可以使用双指针技术来优化。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "medium",
        time_limit: 2000,
        memory_limit: 256,
        tags: ["数组", "双指针", "排序"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher1._id,
        created_at: new Date(Date.now() - 86400000 * 17),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "最长回文子串",
        description: "给你一个字符串 s，找到 s 中最长的回文子串。",
        input: "一个字符串 s",
        output: "输出最长的回文子串",
        sample_input: "babad",
        sample_output: "bab",
        hint: "可以使用中心扩展算法或动态规划。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "medium",
        time_limit: 2000,
        memory_limit: 256,
        tags: ["字符串", "动态规划"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher2._id,
        created_at: new Date(Date.now() - 86400000 * 16),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "正则表达式匹配",
        description: "给你一个字符串 s 和一个字符规律 p，请你来实现一个支持 '.' 和 '*' 的正则表达式匹配。\n\n'.' 匹配任意单个字符\n'*' 匹配零个或多个前面的那一个元素",
        input: "第一行包含字符串 s\n第二行包含模式 p",
        output: "如果匹配成功输出 true，否则输出 false",
        sample_input: "aa\na*",
        sample_output: "true",
        hint: "这是一个经典的动态规划问题。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "hard",
        time_limit: 3000,
        memory_limit: 512,
        tags: ["字符串", "动态规划", "递归"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher1._id,
        created_at: new Date(Date.now() - 86400000 * 15),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "合并K个升序链表",
        description: "给你一个链表数组，每个链表都已经按升序排列。\n\n请你将所有链表合并到一个升序链表中，返回合并后的链表。",
        input: "第一行包含一个整数 k，表示链表数量。\n接下来 k 行，每行表示一个链表的节点值。",
        output: "输出合并后的链表节点值。",
        sample_input: "3\n1 4 5\n1 3 4\n2 6",
        sample_output: "1 1 2 3 4 4 5 6",
        hint: "可以使用分治法或优先队列。",
        source: "LeetCode",
        author: "LeetCode",
        difficulty: "hard",
        time_limit: 3000,
        memory_limit: 512,
        tags: ["链表", "分治", "堆"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher2._id,
        created_at: new Date(Date.now() - 86400000 * 14),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        title: "斐波那契数列",
        description: "斐波那契数列是这样一个数列：0, 1, 1, 2, 3, 5, 8, 13, 21, 34, ...\n\n即 F(0) = 0, F(1) = 1, F(n) = F(n-1) + F(n-2) (n ≥ 2)。\n\n给定 n，求 F(n)。",
        input: "一个非负整数 n (0 ≤ n ≤ 50)",
        output: "输出 F(n) 的值",
        sample_input: "10",
        sample_output: "55",
        hint: "注意避免重复计算，可以使用动态规划。",
        source: "经典算法",
        author: "张教授",
        difficulty: "easy",
        time_limit: 1000,
        memory_limit: 256,
        tags: ["动态规划", "数学"],
        status: "published",
        is_public: true,
        ac_count: 0,
        submit_count: 0,
        created_by: teacher1._id,
        created_at: new Date(Date.now() - 86400000 * 13),
        updated_at: new Date()
    }
];

db.problems.insertMany(problems);
print(`✅ 插入 ${problems.length} 个题目`);

// ==================== 3. 测试用例数据 ====================
print('🧪 生成测试用例数据...');

const testCases = [];
const problemsArray = db.problems.find().toArray();

// 为每个题目生成测试用例
problemsArray.forEach((problem, index) => {
    // 示例测试用例
    testCases.push({
        _id: ObjectId(),
        problem_id: problem._id,
        input: problem.sample_input,
        output: problem.sample_output,
        is_sample: true,
        score: 0,
        created_at: new Date()
    });
    
    // 根据题目生成额外的测试用例
    switch (index) {
        case 0: // 两数之和
            testCases.push(
                {
                    _id: ObjectId(),
                    problem_id: problem._id,
                    input: "3\n3 2 4\n6",
                    output: "1 2",
                    is_sample: false,
                    score: 50,
                    created_at: new Date()
                },
                {
                    _id: ObjectId(),
                    problem_id: problem._id,
                    input: "2\n3 3\n6",
                    output: "0 1",
                    is_sample: false,
                    score: 50,
                    created_at: new Date()
                }
            );
            break;
        case 1: // 回文数
            testCases.push(
                {
                    _id: ObjectId(),
                    problem_id: problem._id,
                    input: "-121",
                    output: "false",
                    is_sample: false,
                    score: 33,
                    created_at: new Date()
                },
                {
                    _id: ObjectId(),
                    problem_id: problem._id,
                    input: "10",
                    output: "false",
                    is_sample: false,
                    score: 33,
                    created_at: new Date()
                },
                {
                    _id: ObjectId(),
                    problem_id: problem._id,
                    input: "1221",
                    output: "true",
                    is_sample: false,
                    score: 34,
                    created_at: new Date()
                }
            );
            break;
        default:
            // 【禁止生成占位用例】之前这里写的是 input:"测试输入1" / output:"测试输出1"，
            // 后果：任何从标准输入读数据的程序都会直接崩溃——例如 Java 的
            // Scanner.nextInt() 读到"测试输入1"抛 InputMismatchException，
            // 判题结果显示"运行时错误"，而学生在页面上"运行示例"却完全正常
            // （运行只用 sample_input，提交才跑全部用例）。
            // 缺用例的题目请在上面的 switch 里补齐真实输入输出；
            // 未覆盖的题目这里只告警、不再插入垃圾数据。
            print(`⚠️  题目「${problem.title}」未预置测试用例，已跳过（请手工补充真实 input/output）`);
            break;

    }
});

db.test_cases.insertMany(testCases);
print(`✅ 插入 ${testCases.length} 个测试用例`);

// ==================== 4. 课程数据 ====================
print('📖 生成课程数据...');

const courses = [
    {
        _id: ObjectId(),
        name: "算法与数据结构",
        avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=algorithm",
        description: "本课程主要介绍常用的数据结构和算法，包括线性表、栈、队列、树、图等数据结构，以及排序、查找、动态规划等算法。",
        teacher_ids: [teacher1._id],
        created_by: teacher1._id,
        // 修复：使用整数类型而不是字符串
        status: 1,  // 1: CourseStatusActive
        ctime: new Date(Date.now() - 86400000 * 30),
        mtime: new Date()
    },
    {
        _id: ObjectId(),
        name: "程序设计基础",
        avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=programming",
        description: "程序设计基础课程，主要学习C++编程语言的基本语法、面向对象编程思想，以及基本的程序设计方法。",
        teacher_ids: [teacher2._id],
        created_by: teacher2._id,
        // 修复：使用整数类型而不是字符串
        status: 1,  // 1: CourseStatusActive
        ctime: new Date(Date.now() - 86400000 * 25),
        mtime: new Date()
    },
    {
        _id: ObjectId(),
        name: "高级算法设计",
        avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=advanced",
        description: "高级算法设计与分析，包括贪心算法、分治算法、动态规划、网络流、字符串算法等高级主题。",
        teacher_ids: [teacher1._id, teacher2._id],
        created_by: teacher1._id,
        // 修复：使用整数类型而不是字符串
        status: 1,  // 1: CourseStatusActive
        ctime: new Date(Date.now() - 86400000 * 20),
        mtime: new Date()
    }
];

db.courses.insertMany(courses);
print(`✅ 插入 ${courses.length} 个课程`);

// ==================== 5. 班级数据 ====================
print('👥 生成班级数据...');

const coursesArray = db.courses.find().toArray();
const clazzes = [];

coursesArray.forEach((course, index) => {
    // 为每个课程创建1-2个班级
    const classCount = index === 2 ? 1 : 2; // 高级算法只有1个班
    
    for (let i = 0; i < classCount; i++) {
        const memberIds = students.slice(i * 10, (i + 1) * 10).map(s => s._id);
        
        clazzes.push({
            _id: ObjectId(),
            name: `${course.name} - 第${i + 1}班`,
            description: `${course.name}课程的第${i + 1}个教学班`,
            course_id: course._id,
            schedule: `周${i + 1 === 1 ? '二' : '四'} 14:00-16:00`,
            member_ids: memberIds,
            teacher_ids: course.teacher_ids,
            require_invite: false,
            max_members: 50,
            add_nums: memberIds.length,
            // 修复：使用整数类型而不是字符串
            status: 1,  // 1: ClassStatusActive
            ctime: course.ctime,
            c_id: course.created_by,
            mtime: new Date()
        });
    }
});

db.clazzes.insertMany(clazzes);
print(`✅ 插入 ${clazzes.length} 个班级`);

// ==================== 6. 课程任务数据 ====================
print('📋 生成课程任务数据...');

const tasks = [];
const clazzesArray = db.clazzes.find().toArray();

clazzesArray.forEach((clazz, clazzIndex) => {
    const course = coursesArray.find(c => c._id.equals(clazz.course_id));
    
    // 为每个班级创建3-5个任务
    const taskCount = 3 + (clazzIndex % 3);
    
    for (let i = 0; i < taskCount; i++) {
        const startTime = new Date(Date.now() - 86400000 * (20 - i * 3));
        const endTime = new Date(startTime.getTime() + 86400000 * 7); // 7天后截止
        
        // 随机选择一些题目作为任务
        const taskProblems = problemsArray.slice(i, i + 2).map(p => p._id);
        
        // 随机一些学生完成任务
        const finishCount = Math.floor(Math.random() * clazz.member_ids.length * 0.7);
        const finishIds = clazz.member_ids.slice(0, finishCount);
        
        tasks.push({
            _id: ObjectId(),
            title: `第${i + 1}周编程作业`,
            description: `本周需要完成${taskProblems.length}道编程题目，请认真阅读题目要求并提交代码。`,
            type: 1,  
            start_time: startTime,
            end_time: endTime,
            relation_ids: taskProblems,
            // 修复：使用整数类型而不是字符串
            status: Date.now() > endTime.getTime() ? 2 : 1,  // 2: TaskStatusEnded, 1: TaskStatusActive
            finish_ids: finishIds,
            course_id: course._id,
            clazz_id: clazz._id,
            ctime: startTime,
            c_id: course.created_by,
            mtime: new Date()
        });
    }
});

db.tasks.insertMany(tasks);
print(`✅ 插入 ${tasks.length} 个课程任务`);

// ==================== 7. 提交数据 ====================
print('💻 生成提交数据...');

const submits = [];
const languages = ["cpp", "java", "python", "c"];
const statuses = ["accepted", "wrong_answer", "time_limit", "compile_error", "runtime_error"];

// 为每个学生生成一些提交记录
students.forEach((student, studentIndex) => {
    const submitCount = 5 + Math.floor(Math.random() * 10); // 每个学生5-15个提交
    
    for (let i = 0; i < submitCount; i++) {
        const problem = problemsArray[Math.floor(Math.random() * problemsArray.length)];
        const language = languages[Math.floor(Math.random() * languages.length)];
        const status = statuses[Math.floor(Math.random() * statuses.length)];
        
        // 生成示例代码
        let code = "";
        switch (language) {
            case "cpp":
                code = `#include <iostream>
#include <vector>
using namespace std;

int main() {
    // 学生${studentIndex + 1}的解答
    cout << "Hello World!" << endl;
    return 0;
}`;
                break;
            case "java":
                code = `public class Solution {
    public static void main(String[] args) {
        // 学生${studentIndex + 1}的解答
        System.out.println("Hello World!");
    }
}`;
                break;
            case "python":
                code = `# 学生${studentIndex + 1}的解答
def main():
    print("Hello World!")

if __name__ == "__main__":
    main()`;
                break;
            case "c":
                code = `#include <stdio.h>

int main() {
    // 学生${studentIndex + 1}的解答
    printf("Hello World!\\n");
    return 0;
}`;
                break;
        }
        
        const submitTime = new Date(Date.now() - Math.random() * 86400000 * 15); // 15天内随机时间
        
        const submit = {
            _id: ObjectId(),
            problem_id: problem._id,
            user_id: student._id,
            code: code,
            language: language,
            status: status,
            created_at: submitTime,
            updated_at: submitTime
        };
        
        // 如果状态不是pending或running，添加判题结果
        if (status !== "pending" && status !== "running") {
            submit.result = {
                status: status,
                time_used: Math.floor(Math.random() * 1000) + 100, // 100-1100ms
                memory_used: Math.floor(Math.random() * 50000) + 10000, // 10-60MB
                compile_error: status === "compile_error" ? "编译错误：语法错误" : "",
                runtime_error: status === "runtime_error" ? "运行时错误：数组越界" : "",
                test_results: []
            };
        }
        
        submits.push(submit);
    }
});

db.submits.insertMany(submits);
print(`✅ 插入 ${submits.length} 个提交记录`);

// ==================== 8. 更新题目统计 ====================
print('📊 更新题目统计信息...');

problemsArray.forEach(problem => {
    const problemSubmits = db.submits.find({problem_id: problem._id}).toArray();
    const acCount = problemSubmits.filter(s => s.status === "accepted").length;
    
    db.problems.updateOne(
        {_id: problem._id},
        {
            $set: {
                submit_count: problemSubmits.length,
                ac_count: acCount,
                updated_at: new Date()
            }
        }
    );
});

print('✅ 题目统计信息更新完成');

// ==================== 9. 生成关系数据 ====================
print('🔗 生成用户任务关系数据...');

const relationUsers = [];
const tasksArray = db.tasks.find().toArray();

tasksArray.forEach(task => {
    task.relation_ids.forEach(relationId => {
        task.finish_ids.forEach(userId => {
            relationUsers.push({
                _id: ObjectId(),
                task_id: task._id,
                relation_id: relationId,
                user_id: userId,
                ctime: new Date(task.ctime.getTime() + Math.random() * 86400000 * 3) // 任务开始后3天内完成
            });
        });
    });
});

if (relationUsers.length > 0) {
    db.relations_users.insertMany(relationUsers);
    print(`✅ 插入 ${relationUsers.length} 个用户任务关系`);
}

// ==================== 10. 创建索引 ====================
print('🔍 创建数据库索引...');

// 用户索引
db.users.createIndex({username: 1}, {unique: true});
db.users.createIndex({email: 1}, {unique: true});
db.users.createIndex({student_id: 1});
db.users.createIndex({role: 1});

// 题目索引
db.problems.createIndex({title: 1});
db.problems.createIndex({difficulty: 1});
db.problems.createIndex({status: 1});
db.problems.createIndex({is_public: 1});
db.problems.createIndex({created_by: 1});
db.problems.createIndex({tags: 1});

// 提交索引
db.submits.createIndex({user_id: 1, problem_id: 1, status: 1});
db.submits.createIndex({problem_id: 1});
db.submits.createIndex({user_id: 1});
db.submits.createIndex({status: 1});
db.submits.createIndex({created_at: -1});

// 测试用例索引
db.test_cases.createIndex({problem_id: 1});
db.test_cases.createIndex({is_sample: 1});

// 课程索引
db.courses.createIndex({created_by: 1});
db.courses.createIndex({status: 1});
db.courses.createIndex({teacher_ids: 1});

// 班级索引
db.clazzes.createIndex({course_id: 1});
db.clazzes.createIndex({teacher_ids: 1});
db.clazzes.createIndex({member_ids: 1});

// 任务索引
db.tasks.createIndex({course_id: 1});
db.tasks.createIndex({clazz_id: 1});
db.tasks.createIndex({status: 1});
db.tasks.createIndex({start_time: 1});
db.tasks.createIndex({end_time: 1});

// 关系索引
db.relations_users.createIndex({task_id: 1});
db.relations_users.createIndex({user_id: 1});
db.relations_users.createIndex({relation_id: 1});

print('✅ 数据库索引创建完成');

// ==================== 数据生成完成 ====================
print('🎉 测试数据生成完成！');
print('');
print('📈 数据统计:');
print(`👤 用户: ${db.users.count()} 个 (1个管理员, 2个教师, ${db.users.count({role: "student"})} 个学生)`);
print(`📚 题目: ${db.problems.count()} 个 (${db.problems.count({difficulty: "easy"})} 简单, ${db.problems.count({difficulty: "medium"})} 中等, ${db.problems.count({difficulty: "hard"})} 困难)`);
print(`🧪 测试用例: ${db.test_cases.count()} 个`);
print(`📖 课程: ${db.courses.count()} 个`);
print(`👥 班级: ${db.clazzes.count()} 个`);
print(`📋 任务: ${db.tasks.count()} 个`);
print(`💻 提交: ${db.submits.count()} 个`);
print(`🔗 关系: ${db.relations_users.count()} 个`);
print('');
print('🔑 默认登录信息:');
print('管理员: admin / admin123');
print('教师1: teacher1 / admin123');
print('教师2: teacher2 / admin123');
print('学生: student1-student20 / admin123');
print('');
print('🌐 现在可以启动服务并访问 API 文档:');
print('- Scalar文档: http://localhost:8080/docs');
print('- Swagger文档: http://localhost:8080/swagger/index.html');
