import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Тип нагрузки
export const options = {
    vus: 300,
    duration: '10s',
};

const authorIds = [
    'bd0768a8-6dea-4e78-936c-e4f6d44a94d3',
    '9671ee22-8ab3-4fd2-93f3-2e6e8eb8cbd8',
    '68ce72ad-5e25-4db0-a3ab-3840519ec31e',
    '6541a244-6b43-4d5e-8c12-913b27eebf4e',
    '9f4d696d-daad-4e9a-95b5-9051c5791858',
];

export default function () {
    const url = 'http://localhost:8080/v1/library/book';

    const randomAuthorId = randomItem(authorIds);

    const payload = JSON.stringify({
        name: `book-${__VU}-${__ITER}`,
        author_id: [randomAuthorId],
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    });

    sleep(0.1);
}