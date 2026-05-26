<script>
	export default {
		onLaunch: function() {
			console.log('App Launch')
		},
		onShow: function() {
			console.log('App Show')
		},
		onHide: function() {
			console.log('App Hide')
		}
	}
</script>

<style>
	/* 每个页面公共 css */
	page,
	:root {
		--sun-primary: #ffb84d;
		--sun-action: #ff7a45;
		--sun-sky: #4b9fff;
		--sun-mint: #63c784;
		--sun-rose: #ef6b7a;
		--sun-ink: #2b2a28;
		--sun-muted: #7b8190;
		--sun-line: #edf0f5;
		--sun-paper: #ffffff;
		--sun-warm: #fff8ea;
	}

	button {
		box-sizing: border-box;
		min-height: 72rpx;
		margin: 0;
		border: 0;
		border-radius: 999rpx;
		background: var(--sun-action);
		color: #fff;
		font-size: 26rpx;
		font-weight: 800;
		line-height: 72rpx;
		padding: 0 28rpx;
		transition: transform 0.16s ease, opacity 0.16s ease, box-shadow 0.16s ease;
		box-shadow: 0 10rpx 22rpx rgba(255, 122, 69, 0.2);
	}

	button::after {
		border: 0;
	}

	button:active {
		transform: scale(0.98);
		opacity: 0.92;
	}

	button[disabled] {
		background: #edf1f6;
		color: #9aa3b2;
		box-shadow: none;
		opacity: 1;
	}

	.sun-page {
		min-height: 100vh;
		padding: 28rpx 28rpx 160rpx;
		background: linear-gradient(180deg, #fff7e7 0%, #eef7ff 42%, #f8fbff 100%);
		color: var(--sun-ink);
	}

	.sun-title {
		display: block;
		font-size: 46rpx;
		line-height: 1.18;
		font-weight: 800;
		color: var(--sun-ink);
	}

	.sun-subtitle {
		display: block;
		margin-top: 10rpx;
		font-size: 24rpx;
		line-height: 1.55;
		color: var(--sun-muted);
	}

	.sun-card {
		margin-bottom: 22rpx;
		padding: 26rpx;
		border: 1rpx solid rgba(255, 184, 77, 0.28);
		border-radius: 28rpx;
		background: rgba(255, 255, 255, 0.9);
		box-shadow: 0 14rpx 34rpx rgba(43, 42, 40, 0.08);
	}

	.sun-card-title {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16rpx;
		font-size: 28rpx;
		font-weight: 800;
		color: var(--sun-ink);
	}

	.sun-btn {
		min-height: 72rpx;
		border-radius: 999rpx;
		padding: 0 28rpx;
		background: var(--sun-action);
		color: #fff;
		font-size: 26rpx;
		font-weight: 800;
		line-height: 72rpx;
		box-shadow: 0 10rpx 22rpx rgba(255, 122, 69, 0.2);
	}

	.sun-btn.secondary {
		background: #eef7ff;
		color: var(--sun-sky);
		border: 1rpx solid rgba(75, 159, 255, 0.22);
	}

	.sun-btn.danger {
		background: #fff1f3;
		color: #d95061;
		border: 1rpx solid rgba(239, 107, 122, 0.2);
	}

	.sun-chip {
		flex: 0 0 auto;
		padding: 14rpx 22rpx;
		border-radius: 999rpx;
		background: rgba(255, 255, 255, 0.9);
		border: 1rpx solid var(--sun-line);
		color: #697180;
		font-size: 22rpx;
		font-weight: 700;
	}

	.sun-chip.active {
		color: #fff;
		background: var(--sun-sky);
		border-color: var(--sun-sky);
	}
</style>
