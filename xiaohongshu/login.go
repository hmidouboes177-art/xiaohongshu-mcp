package xiaohongshu

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

const exploreURL = "https://www.xiaohongshu.com/explore"

type LoginAction struct {
	page *rod.Page
}

func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx)
	if err := navigateExplore(pp); err != nil {
		return false, err
	}

	time.Sleep(1 * time.Second)

	exists, _, err := pp.Has(`.main-container .user .link-wrapper .channel`)
	if err != nil {
		return false, errors.Wrap(err, "check login status failed")
	}

	if !exists {
		return false, nil
	}

	return true, nil
}

func (a *LoginAction) Login(ctx context.Context) error {
	pp := a.page.Context(ctx)

	// 导航到小红书首页，这会触发二维码弹窗
	if err := navigateExplore(pp); err != nil {
		return err
	}

	// 等待一小段时间让页面完全加载
	time.Sleep(2 * time.Second)

	// 检查是否已经登录
	if exists, _, _ := pp.Has(".main-container .user .link-wrapper .channel"); exists {
		// 已经登录，直接返回
		return nil
	}

	// 等待扫码成功提示或者登录完成
	// 这里我们等待登录成功的元素出现，这样更简单可靠
	if _, err := pp.Element(".main-container .user .link-wrapper .channel"); err != nil {
		return errors.Wrap(err, "wait login success element failed")
	}

	return nil
}

func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx)

	// 导航到小红书首页，这会触发二维码弹窗
	if err := navigateExplore(pp); err != nil {
		return "", false, err
	}

	// 等待一小段时间让页面完全加载
	time.Sleep(2 * time.Second)

	// 检查是否已经登录
	if exists, _, _ := pp.Has(".main-container .user .link-wrapper .channel"); exists {
		return "", true, nil
	}

	// 获取二维码图片
	qrcode, err := pp.Element(".login-container .qrcode-img")
	if err != nil {
		return "", false, errors.Wrap(err, "get qrcode element failed")
	}

	src, err := qrcode.Attribute("src")
	if err != nil {
		return "", false, errors.Wrap(err, "get qrcode src failed")
	}
	if src == nil || len(*src) == 0 {
		return "", false, errors.New("qrcode src is empty")
	}

	return *src, false, nil
}

func navigateExplore(page *rod.Page) error {
	if err := page.Navigate(exploreURL); err != nil {
		return errors.Wrap(err, "navigate to xiaohongshu failed, please check proxy credentials/network")
	}

	if err := page.WaitLoad(); err != nil {
		return errors.Wrap(err, "wait xiaohongshu page load failed")
	}

	return nil
}

func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			el, err := pp.Element(".main-container .user .link-wrapper .channel")
			if err == nil && el != nil {
				return true
			}
		}
	}
}
