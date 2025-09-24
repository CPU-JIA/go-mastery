package utils

import (
	"fmt"
	"time"
)

// FormatDuration 格式化时间间隔为人类可读的字符串
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f小时", d.Hours())
	}
	return fmt.Sprintf("%.1f天", d.Hours()/24)
}

// GetChineseWeekday 获取中文星期
func GetChineseWeekday(t time.Time) string {
	weekdays := []string{
		"星期日", "星期一", "星期二", "星期三",
		"星期四", "星期五", "星期六",
	}
	return weekdays[t.Weekday()]
}

// IsWorkday 判断是否为工作日（周一到周五）
func IsWorkday(t time.Time) bool {
	weekday := t.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}

// GetMonthStart 获取月份开始时间
func GetMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetMonthEnd 获取月份结束时间
func GetMonthEnd(t time.Time) time.Time {
	return GetMonthStart(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// GetWeekStart 获取一周开始时间（星期一）
func GetWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	days := int(weekday) - 1
	return time.Date(t.Year(), t.Month(), t.Day()-days, 0, 0, 0, 0, t.Location())
}

// GetWeekEnd 获取一周结束时间（星期日）
func GetWeekEnd(t time.Time) time.Time {
	return GetWeekStart(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// ParseChineseDate 解析中文日期格式
func ParseChineseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006年01月02日",
		"2006年1月2日",
		"2006-01-02",
		"2006/01/02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}

// FormatChineseDate 格式化为中文日期
func FormatChineseDate(t time.Time) string {
	return t.Format("2006年01月02日")
}

// TimeAgo 计算时间差并返回人类可读的字符串
func TimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "刚刚"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d小时前", hours)
	}
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	}

	return FormatChineseDate(t)
}

// IsLeapYear 判断是否为闰年
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// GetQuarter 获取季度
func GetQuarter(t time.Time) int {
	month := t.Month()
	return (int(month)-1)/3 + 1
}

// GetAge 根据生日计算年龄
func GetAge(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()

	// 如果今年的生日还没过，年龄减1
	if now.Month() < birthday.Month() ||
		(now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		age--
	}

	return age
}

func init() {
	fmt.Println("utils time 模块初始化")
}
