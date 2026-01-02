const canvas = document.getElementById("bg-canvas");
const ctx = canvas.getContext("2d");

let width, height;
let nodes = [];
const NODE_COUNT = 50;
const MAX_DIST = 140;
const MOUSE_RADIUS = 180;

function resize() {
    width = canvas.width = window.innerWidth;
    height = canvas.height = window.innerHeight;
}
window.addEventListener("resize", resize);
resize();

const mouse = { x: null, y: null };

window.addEventListener("mousemove", e => {
    mouse.x = e.clientX;
    mouse.y = e.clientY;
});

window.addEventListener("mouseleave", () => {
    mouse.x = null;
    mouse.y = null;
});

class Node {
    constructor() {
        this.x = Math.random() * width;
        this.y = Math.random() * height;
        this.vx = (Math.random() - 0.5) * 0.6;
        this.vy = (Math.random() - 0.5) * 0.6;
    }
    move() {
        this.x += this.vx;
        this.y += this.vy;

        if (this.x < 0 || this.x > width) this.vx *= -1;
        if (this.y < 0 || this.y > height) this.vy *= -1;

        // 🔥 mouse attraction (VISIBLE but controlled)
        if (mouse.x !== null) {
            const dx = mouse.x - this.x;
            const dy = mouse.y - this.y;
            const dist = Math.sqrt(dx * dx + dy * dy);

            if (dist < MOUSE_RADIUS) {
                const force = (MOUSE_RADIUS - dist) / MOUSE_RADIUS;
                this.x += dx * force * 0.01;
                this.y += dy * force * 0.01;
            }
        }
    }
    draw() {
        ctx.fillStyle = "rgba(255,255,255,0.45)";
        ctx.beginPath();
        ctx.arc(this.x, this.y, 1.7, 0, Math.PI * 2);
        ctx.fill();
    }
}

function init() {
    nodes = [];
    for (let i = 0; i < NODE_COUNT; i++) {
        nodes.push(new Node());
    }
}
init();

function drawLines() {
    for (let i = 0; i < nodes.length; i++) {
        for (let j = i + 1; j < nodes.length; j++) {
            const dx = nodes[i].x - nodes[j].x;
            const dy = nodes[i].y - nodes[j].y;
            const dist = Math.sqrt(dx * dx + dy * dy);

            if (dist < MAX_DIST) {
                let opacity = 0.08 * (1 - dist / MAX_DIST);

                // 🔥 highlight connections near mouse
                if (mouse.x !== null) {
                    const mx = (nodes[i].x + nodes[j].x) / 2 - mouse.x;
                    const my = (nodes[i].y + nodes[j].y) / 2 - mouse.y;
                    const md = Math.sqrt(mx * mx + my * my);
                    if (md < MOUSE_RADIUS) {
                        opacity += 0.12;
                    }
                }

                ctx.strokeStyle = `rgba(255,255,255,${opacity})`;
                ctx.beginPath();
                ctx.moveTo(nodes[i].x, nodes[i].y);
                ctx.lineTo(nodes[j].x, nodes[j].y);
                ctx.stroke();
            }
        }
    }
}

function animate() {
    ctx.clearRect(0, 0, width, height);

    nodes.forEach(n => {
        n.move();
        n.draw();
    });

    drawLines();
    requestAnimationFrame(animate);
}

animate();
