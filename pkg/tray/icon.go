package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// GenerateTAIcon 生成带有 "TA" (Tim-Agent) 字母标识的精致应用与托盘图标
// 支持任意尺寸如 32x32 (托盘) 或 128x128 / 256x256 (应用图标)
func GenerateTAIcon(size int) []byte {
	if size <= 0 {
		size = 32
	}
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// 背景：深色高质感圆角渐变底色 (现代终端蓝黑)
	// (x, y) 处带圆角防溢出
	radius := float64(size) * 0.22
	cx := float64(size) / 2.0
	cy := float64(size) / 2.0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// 圆角矩形遮罩判定
			dx := math.Max(0, math.Abs(float64(x)-cx+0.5)-(cx-radius))
			dy := math.Max(0, math.Abs(float64(y)-cy+0.5)-(cy-radius))
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= radius {
				// 渐变底色：从顶部 #1f2937 (31, 41, 55) 到底部 #0f172a (15, 23, 42)
				ratio := float64(y) / float64(size)
				r := uint8(31 - 16*ratio)
				g := uint8(41 - 18*ratio)
				b := uint8(55 - 13*ratio)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			} else {
				// 透明背景
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	// 绘制发光精致的 "T" 和 "A" 双字母组合
	// "T": 渐变科技青蓝 (#38bdf8), "A": 渐变霓虹绿 (#34d399)
	scale := float64(size) / 32.0

	// 辅助填充矩形函数
	fillRect := func(x0, y0, w, h float64, c color.Color) {
		ix0 := int(math.Floor(x0 * scale))
		iy0 := int(math.Floor(y0 * scale))
		ix1 := int(math.Ceil((x0 + w) * scale))
		iy1 := int(math.Ceil((y0 + h) * scale))
		for py := iy0; py < iy1; py++ {
			for px := ix0; px < ix1; px++ {
				if px >= 0 && px < size && py >= 0 && py < size {
					img.Set(px, py, c)
				}
			}
		}
	}

	colorT := color.RGBA{R: 56, G: 189, B: 248, A: 255} // #38bdf8 亮青
	colorA := color.RGBA{R: 52, G: 211, B: 153, A: 255} // #34d399 翡翠绿

	// 1. 字母 "T": 左侧区域 (x: 5~15)
	// 横杠: x=5~15, y=7~10.5
	fillRect(5, 7, 10, 3.2, colorT)
	// 竖杠: x=8.5~11.5, y=10~25
	fillRect(8.4, 10, 3.2, 15, colorT)

	// 2. 字母 "A": 右侧区域 (x: 16~27)
	// 左斜柱
	for i := 0.0; i <= 15.0; i += 0.5 {
		currX := 21.0 - (i * 0.35)
		currY := 9.0 + i
		fillRect(currX, currY, 2.8, 1.2, colorA)
	}
	// 右斜柱
	for i := 0.0; i <= 15.0; i += 0.5 {
		currX := 22.0 + (i * 0.35)
		currY := 9.0 + i
		fillRect(currX, currY, 2.8, 1.2, colorA)
	}
	// A的中间横杠: y=16~18.5, x=18~25
	fillRect(18.2, 16.5, 7.2, 2.6, colorA)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// GenerateTrayIcon 动态生成 32x32 的精致 "TA" 终端托盘图标
func GenerateTrayIcon() []byte {
	return GenerateTAIcon(32)
}

// GenerateAppIcon 生成 256x256 高清应用图标
func GenerateAppIcon() []byte {
	return GenerateTAIcon(256)
}
