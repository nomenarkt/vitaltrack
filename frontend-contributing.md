# 🧑‍🎨 FRONTEND CONTRIBUTING GUIDELINES  
Welcome! This doc defines how to contribute to our frontend codebase (Web: Next.js, Mobile: React Native). All contributions must follow component-driven architecture, platform-aware design, and test-first development.
## 🧱 Architecture  
Structure is separated into:  
- `/web` – Next.js admin  
- `/mobile` – React Native user  
- `/shared` – platform-neutral logic: types, hooks, schemas, API
Components call hooks, backed by shared logic. Shared code must be fully type-safe and platform-agnostic.
## 🧠 Codex Rules  
- Follow tasks from The Polyglot exactly  
- Do NOT invent props, flows, or UI  
- Use Zod schemas, tokens, and patterns as defined
## ✅ Commits & PRs  
- One PR = one feature or fix  
- Use Conventional Commits:  
  - `feat(web): add RefillCard`  
  - `fix(mobile): correct keyboard overlap`  
- Keep commits atomic and scoped
## 🧪 Testing  
- All components, hooks, and logic must be covered with Jest and RTL (web) or RTL-native (mobile)  
- Test files live beside code (`RefillCard.tsx`, `RefillCard.test.tsx`)  
- Test props, interactions, edge cases, and validation errors
## 🎨 Styling  
- Web: Tailwind CSS  
- Mobile: Dripsy or nativewind  
- Use tokens from `theme.ts` (no raw colors or spacing)
## 📦 Shared Code  
- Business logic, schemas, and types go in `/shared`  
- Must not import from web or mobile directly  
- All shared exports must be tested
## 🚫 DO NOT  
- Add TODOs or commented-out code  
- Use raw fetch/axios in components (wrap in hooks)  
- Mix platform-specific logic into `/shared`  
- Push untested commits
## ✅ YOU MUST  
- Use Zod + react-hook-form for all forms  
- Follow TDD (write tests before merging)  
- Respect a11y: labels, roles, keyboard focus  
- Use suspense + error boundaries where needed
By contributing, you help us build a scalable, testable, and maintainable frontend across web and mobile.