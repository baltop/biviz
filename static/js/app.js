// BiViz - 클라이언트 사이드 스크립트

// HTMX 이벤트 리스너
document.addEventListener('htmx:responseError', function(event) {
    console.error('HTMX 요청 실패:', event.detail);
});

// 페이지 전환 시 부드러운 애니메이션
document.addEventListener('htmx:afterSwap', function(event) {
    // 새로 추가된 요소에 fade-in 효과
    const target = event.detail.target;
    if (target) {
        target.style.opacity = '0';
        requestAnimationFrame(() => {
            target.style.transition = 'opacity 200ms ease-in';
            target.style.opacity = '1';
        });
    }
});
