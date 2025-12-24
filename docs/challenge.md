# Challenge description:

↩️ [Return to README](../README.md)

# PARTE 1: DESAFÍO DE DISEÑO DE
## ARQUITECTURA
Resumen: Diseñar una arquitectura de sistema de procesamiento de pagos que maneje el pago de
servicios, gestión de billetera, recolección de métricas básicas y respuestas de pasarelas de
pago con manejo de errores apropiado.

## Requisitos del Negocio
Funcionalidad Principal

- Procesamiento de Pagos: Los usuarios pueden pagar servicios usando el saldo de su
billetera a través de una API REST.
- Gestión de Billetera: Rastrear saldos de usuarios, deducir fondos, manejar reembolsos
y consultar historial.
- Integración con Pasarela: Comunicación con pasarela de pago externa (mock).
- Manejo de Errores: Gestión adecuada de fallas y validaciones.

## Entregables del Diseño de Arquitectura
1. Diagrama de Arquitectura del Sistema
Crear diagramas que muestren:
- Arquitectura de alto nivel con servicios principales, base de datos y componentes.
- Flujo de solicitudes HTTP/REST para diferentes operaciones.
- Capas de la aplicación (API, Dominio, Infraestructura).
- Puntos de integración con sistemas externos.

2. Documento de Diseño de Servicios
Documentar la estructura del sistema:
- Módulos/Servicios y sus responsabilidades.
- Endpoints de API con métodos HTTP, request/response.
- Modelos de dominio principales (entidades, value objects).
- Separación de capas según principios de arquitectura limpia.

3. Diseño de Base de Datos
Especificar el esquema de base de datos:
- Modelo relacional con tablas, campos y tipos de datos.
- Relaciones entre entidades (foreign keys, índices).
- Manejo de transacciones y consistencia de datos.

4. Recomendación de Stack Tecnológico
Justificar las elecciones tecnológicas para:
- Lenguaje: Go (Golang)
- Base de datos: PostgreSQL u otra relacional
- Framework web: Gin, Echo, Chi, o HTTP estándar de Go
- Contenedores: Docker y docker-compose
- Testing: Framework de testing de Go

5. Estrategia de Manejo de Errores
Gestión de fallas y validaciones:
- Identificación de escenarios de error (saldo insuficiente, usuario no existe, falla de
pasarela).
- Códigos de estado HTTP apropiados.
- Estructura de respuestas de error consistente.
- Validación de entrada y sanitización.
- Logging de errores para debugging.

## Escenarios Específicos a Abordar
Escenarios de Flujo de Pago
1. Ruta Feliz: Pago exitoso de principio a fin.
2. Saldo Insuficiente: Validación y rechazo apropiado.
3. Usuario No Existe: Manejo de entidades no encontradas.
4. Falla de Pasarela Externa: Timeout y manejo de errores externos.
5. Validación de Datos: Montos negativos, campos requeridos, etc.

# PARTE 2: DESAFÍO DE IMPLEMENTACIÓN

## Contexto
Implementar un sistema de pagos basado en la arquitectura diseñada en la Parte 1. El foco
está en demostrar habilidades de diseño, código limpio, buenas prácticas y conocimiento
de Go, no necesariamente en crear una aplicación completamente funcional en producción.
Tiempo estimado: 4-5 horas

## Entregables Requeridos
1. Implementación en Go
- Estructura de proyecto bien organizada siguiendo convenciones de Go
- Código que demuestre aplicación de patrones de diseño
- Separación de capas clara (handlers, services, repositories)
- Interfaces para abstraer dependencias
- Implementación de al menos 3 endpoints principales:
  - POST /wallets/{user_id}/payments - Crear pago
  - GET /wallets/{user_id}/balance - Consultar saldo
  - GET /wallets/{user_id}/transactions - Historial de transacciones

2. Aplicación de Principios
Demostrar comprensión de:
- SOLID Principles
  - Single Responsibility
  - Dependency Inversion
  - Interface Segregation
- Clean Architecture / Hexagonal Architecture
  - Separación entre dominio, aplicación e infraestructura
  - Independencia de frameworks y librerías externas
- Domain-Driven Design (DDD) básico
  - Entidades de dominio
  - Servicios
  - Repositorios
- Clean Code
  - Nombres descriptivos
  - Funciones pequeñas y enfocadas
  - Comentarios solo cuando añaden valor

3. Base de Datos
- Schema SQL con definición de tablas
- Migraciones (puede ser un script SQL simple)
- Uso de transacciones donde sea necesario

4. Testing
- Pruebas unitarias de servicios principales
- Mocks para dependencias externas (pasarela de pago, base de datos)
- Cobertura mínima del 60% en componentes críticos

5. Infraestructura (Opcional pero valorado)
- Dockerfile para containerizar la aplicación,
- Variables de entorno para configuración
- Instrucciones claras de cómo ejecutar el proyecto

6. CI/CD Básico (Opcional pero valorado)
- GitHub Actions o similar con pipeline básico:
  - Lint (golangci-lint)
  - Tests
  - Build

Componentes Opcionales (Puntos Extra)
- Swagger/OpenAPI documentation
- Graceful shutdown
- Context propagation apropiado
- Validación con tags (validator library)
- Logging estructurado (usando slog, zerolog o similar)


Notas Importantes
- No se requiere despliegue en cloud ni funcionalidad 100% operativa
- Usar mocks/simulaciones para la pasarela de pago externa
- Enfocarse en demostrar comprensión de principios y buenas prácticas
- El código debe ser production-like aunque no esté desplegado
- Si usas IA (Cursor, Windsurf, Copilot, etc.), incluir historial de conversación o
mencionar cómo fue utilizada
- Priorizar calidad sobre cantidad de features

Entrega
1. Subir código a repositorio Git público (GitHub, GitLab)
2. README.md detallado que incluya:
- Descripción del proyecto
  - Decisiones arquitectónicas y por qué
  - Stack tecnológico usado
  - Instrucciones de instalación y ejecución
  - Cómo ejecutar tests
  - Endpoints disponibles con ejemplos
  - Mejoras futuras identificadas
3. Compartir el link del repositorio

Challenge original doc download: [semi-sr-backend-challenge.pdf](https://github.com/user-attachments/files/24297353/semi-sr-backend-challenge.pdf)